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
		server := viper.GetViper().GetString(hostGRPC)
		if server == "" {
			log.Fatal(msgErrMissingGRPCServer)
		}
		token := viper.GetViper().GetString(tokenJWT)
		if token == "" {
			log.Fatal(msgErrMissingToken)
		}

		key := viper.GetViper().GetString("secretkey")
		if key == "" {
			log.Fatal("Missing secretkey, update configuration file.")
		}

		dbpath, _ := cmd.Flags().GetString("dbpath")
		if dbpath == "" {
			log.Fatal(msgErrNoDBPath)
		}

		store, err := storage.NewSQLiteDB(dbpath)
		if err != nil {
			log.Fatal(msgErrInitGRPC, err)
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
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
