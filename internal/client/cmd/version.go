package gkcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gkclient",
	Long:  `All software has versions. This is gkclient's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Build version: %s\n", BuildVersion)
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Build commit: %s\n", BuildCommit)
	},
}
