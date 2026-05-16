package main

import (
	"github.com/rajatjindal/krew-release-bot/pkg/source/actions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var dryRun bool

func init() {
	rootCmd.AddCommand(actionCmd)
	actionCmd.Flags().BoolVar(&dryRun, "dry-run", false, "render the release request and skip webhook or PR submission")
}

// actionCmd is the github action command
var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "github action for updating plugin manifests in krew-index repo",
	Run: func(cmd *cobra.Command, args []string) {
		err := actions.RunActionWithOptions(actions.RunOptions{
			DryRun: &dryRun,
		})
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
