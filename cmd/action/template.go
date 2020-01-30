package main

import (
	"fmt"
	"os"

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

	templateCmd.Flags().StringVar(&tagName, "tag", "", "tag name to use for templating")
	templateCmd.MarkFlagRequired("tag")

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
		if err == nil {
			fmt.Println(string(spec))
			os.Exit(0)
		}

		if invalidSpecError, ok := err.(source.InvalidPluginSpecError); ok {
			fmt.Println(invalidSpecError.Spec)
			logrus.Fatal(invalidSpecError.Error())
		}

		logrus.Fatal(err)
	},
}
