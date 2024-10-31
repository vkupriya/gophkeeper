package secret

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
	"github.com/vkupriya/gophkeeper/internal/client/storage"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync secrets with Gophkeeper server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetViper().GetString("token")
		if token == "" {
			fmt.Printf("Missing token, please, login.")
		}

		key := viper.GetViper().GetString("secretkey")
		if key == "" {
			fmt.Printf("Missing secretkey, update configuration file.")
		}

		store, err := storage.NewSQLiteDB()
		if err != nil {
			log.Fatal("Error in setting up DB: ", err)
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc); err != nil {
			log.Fatal("error initializing GRPC client: ", err)
		}

		secretsRemote, err := svc.ListSecrets(token)
		if err != nil {
			log.Fatal("error getting list of secrets: ", err)
		}

		// dropping all stored secrets in local DB
		err = store.SecretDeleteAll()
		if err != nil {
			log.Fatal("error deleting secrets in local database before sync: ", err)
		}

		for _, secret := range secretsRemote {
			remoteItem, err := svc.GetSecret(token, key, secret.Name)
			if err != nil {
				log.Fatal("error getting secret: ", err)
			}
			// Inserting secret into local DB
			err = store.SecretAdd(remoteItem)
			if err != nil {
				log.Fatal("error inserting secret into local DB: ", err)
			}
		}
		fmt.Println("successfully synchronised secret db.")
	},
}

func init() {
	SyncCmd.Flags().StringP("server", "s", "", "GophKeeper server.")
}
