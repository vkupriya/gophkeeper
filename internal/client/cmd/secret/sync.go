package secret

import (
	"fmt"

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
		var msg string
		server := viper.GetViper().GetString(hostGRPC)
		if server == "" {
			cobra.CheckErr(msgErrMissingGRPCServer)
		}
		token := viper.GetViper().GetString(tokenJWT)
		if token == "" {
			cobra.CheckErr(msgErrMissingToken)
		}

		key := viper.GetViper().GetString("secretkey")
		if key == "" {
			cobra.CheckErr("Missing secretkey, update configuration file.")
		}

		dbpath, _ := cmd.Flags().GetString("dbpath")
		if dbpath == "" {
			cobra.CheckErr(msgErrNoDBPath)
		}

		store, err := storage.NewSQLiteDB(dbpath)
		if err != nil {
			msg = fmt.Sprint(msgErrInitGRPC, err)
			cobra.CheckErr(msg)
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			msg = fmt.Sprintf("error initializing GRPC client: %v ", err)
			cobra.CheckErr(msg)
		}

		secretsRemote, err := svc.ListSecrets(token)
		if err != nil {
			msg = fmt.Sprintf("error getting list of secrets: %v", err)
			cobra.CheckErr(msg)
		}

		// dropping all stored secrets in local DB
		err = store.SecretDeleteAll()
		if err != nil {
			msg = fmt.Sprintf("error deleting secrets in local database before sync: %v", err)
			cobra.CheckErr(msg)
		}

		for _, secret := range secretsRemote {
			remoteItem, err := svc.GetSecret(token, key, secret.Name)
			if err != nil {
				msg = fmt.Sprintf("error getting secret: %v ", err)
				cobra.CheckErr(msg)
			}
			// Inserting secret into local DB
			err = store.SecretAdd(remoteItem)
			if err != nil {
				msg = fmt.Sprintf("error inserting secret into local DB: %v", err)
				cobra.CheckErr(msg)
			}
		}
		fmt.Println("successfully synchronised secret db.")
	},
}
