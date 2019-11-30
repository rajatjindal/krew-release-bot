package krew

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty"
	"github.com/sirupsen/logrus"
)

func downloadFileWithName(uri, name string) (string, error) {
	client := resty.New()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	file := filepath.Join(dir, name)
	resp, err := client.R().SetOutput(file).Get(uri)
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("received response-code %d from %s", resp.StatusCode(), uri)
	}

	logrus.Infof("downloaded file %s", file)
	return file, nil
}

func downloadFile(uri string) (string, error) {
	return downloadFileWithName(uri, fmt.Sprintf("%d", time.Now().Unix()))
}

func getSha256ForAsset(uri string) (string, error) {
	file, err := downloadFile(uri)
	if err != nil {
		return "", err
	}

	defer os.Remove(file)
	sha256, err := getSha256(file)
	if err != nil {
		return "", err
	}

	return sha256, nil
}

func getSha256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
