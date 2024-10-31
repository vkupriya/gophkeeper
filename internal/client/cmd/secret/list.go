package secret

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
	"github.com/vkupriya/gophkeeper/internal/client/models"
	"github.com/vkupriya/gophkeeper/internal/client/storage"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list secrets",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var secrets []*models.SecretItem
		token := viper.GetViper().GetString("token")
		if token == "" {
			fmt.Printf("Missing token, please, login.")
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc); err != nil {
			fmt.Println("error initializing GRPC client: ", err)
		}

		secrets, err := svc.ListSecrets(token)
		if err != nil {
			if errors.Is(err, grpcclient.ErrServerUnavailable) {
				fmt.Println("server is unavailable, attempting to read secrets from local DB.")
				store, err := storage.NewSQLiteDB()
				if err != nil {
					log.Fatal("Error in setting up DB: ", err)
				}
				defer store.DB.Close()
				secrets, err = store.SecretList()
				if err != nil {
					fmt.Println("failed to read secrets from local DB", err)
				}

			}
			fmt.Println("error getting list of secrets: ", err)
		}

		if len(secrets) != 0 {
			res, _ := json.MarshalIndent(secrets, "", "   ")
			fmt.Println(string(res))
		}
	},
}

func init() {
	AddCmd.Flags().StringP("server", "s", "", "GophKeeper server.")
}
