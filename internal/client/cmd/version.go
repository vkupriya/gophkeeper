package gkcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gkclient",
	Long:  `All software has versions. This is gkclient's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gkclient version: 0.1")
	},
}
