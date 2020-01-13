package main

import (
	"github.com/rajatjindal/krew-release-bot/pkg/source/actions"
	"github.com/sirupsen/logrus"
)

func main() {
	err := actions.RunAction()
	if err != nil {
		logrus.Fatal(err)
	}
}
