package krew

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

//UpdatePluginManifest updates the manifest with latest release info
func UpdatePluginManifest(templateFileURI, actualFile string, release *github.RepositoryRelease) error {
	templateFile, err := downloadFileWithName(templateFileURI, ".krew.yaml")
	if err != nil {
		return err
	}

	processedPluginBytes, err := processPluginTemplate(templateFile, release)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(actualFile, processedPluginBytes, 0644)
}

func processPluginTemplate(templateFile string, releaseInfo *github.RepositoryRelease) ([]byte, error) {
	t := template.New(".krew.yaml").Funcs(map[string]interface{}{
		"addURIAndSha": func(url, tag string) string {
			t := struct {
				TagName string
			}{
				TagName: tag,
			}
			buf := new(bytes.Buffer)
			temp, err := template.New("url").Parse(url)
			if err != nil {
				panic(err)
			}

			err = temp.Execute(buf, t)
			if err != nil {
				panic(err)
			}

			logrus.Infof("getting sha256 for %s", buf.String())
			sha256, err := getSha256ForAsset(buf.String())
			if err != nil {
				panic(err)
			}

			return fmt.Sprintf(`uri: %s
    sha256: %s`, buf.String(), sha256)
		},
	})

	templateObject, err := t.ParseFiles(templateFile)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = templateObject.Execute(buf, releaseInfo)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

//PluginFileName returns the plugin file with extension
func PluginFileName(name string) string {
	return fmt.Sprintf("%s%s", name, ".yaml")
}
