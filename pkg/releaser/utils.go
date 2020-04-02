package releaser

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
)

//Release releases
func (releaser *Releaser) Release(request *source.ReleaseRequest) (string, error) {
	tempdir, err := ioutil.TempDir("", "krew-index-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempdir)

	logrus.Infof("will operate in tempdir %s", tempdir)
	repo, err := releaser.cloneRepos(tempdir, request)
	if err != nil {
		return "", err
	}

	newIndexFile, err := ioutil.TempFile("", "krew-")
	if err != nil {
		return "", err
	}
	defer os.Remove(newIndexFile.Name())

	err = ioutil.WriteFile(newIndexFile.Name(), request.ProcessedTemplate, 0644)
	if err != nil {
		return "", err
	}

	logrus.Info("validating ownership")
	existingIndexFile := filepath.Join(tempdir, "plugins", krew.PluginFileName(request.PluginName))
	err = krew.ValidateOwnership(existingIndexFile, request.PluginOwner)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("plugin %q not found in existing repo. The first release of a new plugin has to be done manually", request.PluginName)
		}

		return "", fmt.Errorf("failed when validating ownership with error: %s", err.Error())
	}

	logrus.Info("update plugin manifest with latest release info")
	err = krew.ValidatePlugin(request.PluginName, newIndexFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed when validating plugin spec with error: %s", err.Error())
	}

	_, err = copyFile(newIndexFile.Name(), existingIndexFile)
	if err != nil {
		return "", fmt.Errorf("failed when copying plugin spec with error: %s", err.Error())
	}

	logrus.Infof("pushing changes to branch %s", *releaser.getBranchName(request))
	commit := commitConfig{
		Msg:        fmt.Sprintf("new version %s of %s", request.TagName, request.PluginName),
		RemoteName: OriginNameLocal,
	}

	err = releaser.addCommitAndPush(repo, commit, request)
	if err != nil {
		return "", err
	}

	logrus.Info("submitting the pr")
	pr, prID, err := releaser.submitPR(request)
	if err != nil {
		return "", err
	}

	logrus.Info("checking if there are outdated PRs")
	// Close outdated PRs before submiting new one
	err = releaser.closeExistingPR(request, prID)
	if err != nil {
		logrus.Warnf("Failed to close outdated PR %d", prID)
	}

	return pr, nil
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
