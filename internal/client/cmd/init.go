package gkcli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vkupriya/gophkeeper/internal/client/storage"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Init command initialises sqlite db.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var msg string
		dbpath, _ := cmd.Flags().GetString("dbpath")
		fmt.Println("DB path: ", dbpath)
		if dbpath == "" {
			cobra.CheckErr("missing local db path")
		}
		Store, err := storage.NewSQLiteDB(dbpath)
		if err != nil {
			msg = fmt.Sprintf("failed setting up DB: %v", err)
			cobra.CheckErr(msg)
		}

		err = storage.RunMigrations(Store)
		if err != nil {
			msg = fmt.Sprintf("failed running migrations: %v", err)
			cobra.CheckErr(msg)
		}
	},
}
