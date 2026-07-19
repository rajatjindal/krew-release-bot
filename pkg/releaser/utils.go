package releaser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
)

// Release releases
func (releaser *Releaser) Release(request *source.ReleaseRequest) (string, error) {
	tempdir, err := os.MkdirTemp("", "krew-index-")
	if err != nil {
		return "", err
	}
	if !releaser.Config.DryRun {
		defer os.RemoveAll(tempdir)
	}

	logrus.Infof("will operate in tempdir %s", tempdir)
	repo, err := releaser.cloneRepos(tempdir, request)
	if err != nil {
		return "", err
	}

	newIndexFile, err := os.CreateTemp("", "krew-")
	if err != nil {
		return "", err
	}
	defer os.Remove(newIndexFile.Name())

	err = os.WriteFile(newIndexFile.Name(), request.ProcessedTemplate, 0644)
	if err != nil {
		return "", err
	}

	existingIndexFile := filepath.Join(tempdir, "plugins", krew.PluginFileName(request.PluginName))
	logrus.Info("update plugin manifest with latest release info")
	err = krew.ValidatePlugin(request.PluginName, newIndexFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed when validating plugin spec with error: %s", err.Error())
	}

	_, err = copyFile(newIndexFile.Name(), existingIndexFile)
	if err != nil {
		return "", fmt.Errorf("failed when copying plugin spec with error: %s", err.Error())
	}

	commit := commitConfig{
		Msg:        fmt.Sprintf("new version %s of %s", request.TagName, request.PluginName),
		RemoteName: OriginNameLocal,
	}

	err = releaser.addCommit(repo, commit)
	if err != nil {
		return "", err
	}

	if releaser.Config.DryRun {
		logrus.Infof("dry run enabled, leaving tempdir intact for inspection: %s", tempdir)
		logrus.Info("dry run enabled, skipping git push and pull request creation")
		return fmt.Sprintf("dry-run: manifest validated, branch created, and commit written locally at %s; skipped push and PR creation", tempdir), nil
	}

	logrus.Infof("pushing changes to branch %s", releaser.getBranchName(request))

	err = releaser.pushCommit(repo, commit, request)
	if err != nil {
		return "", err
	}

	logrus.Info("submitting the pr")
	pr, err := releaser.submitPR(request)
	if err != nil {
		return "", err
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
