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
		Store, err := storage.NewSQLiteDB()
		if err != nil {
			fmt.Println("Error in setting up DB: ", err)
		}

		err = storage.RunMigrations(Store)
		if err != nil {
			fmt.Println("Error running migrations: ", err)
		}
	},
}
