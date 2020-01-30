package main

import (
	"fmt"

	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tagName      string
	templateFile string
	debug        bool
)

func init() {
	rootCmd.AddCommand(templateCmd)

	templateCmd.Flags().StringVar(&tagName, "tag-name", "", "tag name to use for templating")
	templateCmd.MarkFlagRequired("tag-name")

	templateCmd.Flags().StringVar(&templateFile, "template-file", ".krew.yaml", "template file to use for templating")
	templateCmd.MarkFlagRequired("template-file")

	templateCmd.Flags().BoolVar(&debug, "debug", false, "print debug level logs")
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "template helps validate the krew index template file without going through github actions workflow",
	Run: func(cmd *cobra.Command, args []string) {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		releaseRequest := source.ReleaseRequest{
			TagName: tagName,
		}

		_, spec, err := source.ProcessTemplate(templateFile, releaseRequest)
		if err != nil {
			logrus.Fatal(err)
		}

		fmt.Println(string(spec))
	},
}
