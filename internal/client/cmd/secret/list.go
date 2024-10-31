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
		server := viper.GetViper().GetString(hostGRPC)
		if server == "" {
			log.Fatal(msgErrMissingGRPCServer)
		}

		token := viper.GetViper().GetString(tokenJWT)
		if token == "" {
			log.Fatal(msgErrMissingToken)
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			log.Fatal(msgErrInitGRPC, err)
		}

		secrets, err := svc.ListSecrets(token)
		if err != nil {
			if errors.Is(err, grpcclient.ErrServerUnavailable) {
				fmt.Println("server is unavailable, attempting to read secrets from local DB.")
				store, err := storage.NewSQLiteDB()
				if err != nil {
					log.Fatal("Error in setting up DB: ", err)
				}
				defer func() {
					if err := store.DB.Close(); err != nil {
						log.Fatal("failed to close local DB")
					}
				}()

				secrets, err = store.SecretList()
				if err != nil {
					log.Fatal("failed to read secrets from local DB", err)
				}

			} else {
				log.Fatal("error getting list of secrets: ", err)
			}
		}

		if len(secrets) != 0 {
			res, _ := json.MarshalIndent(secrets, "", "   ")
			fmt.Println(string(res))
		}
	},
}
