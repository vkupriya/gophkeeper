package secret

import (
	"errors"
	"fmt"
	"os"

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
		// map for local secrets for DB Sync
		secretsMapLocal := map[string]int64{}
		// map for remote secrets for DB Sync
		secretsMapRemote := map[string]int64{}
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
			cobra.CheckErr("missing secretkey, update configuration file.")
		}

		dbpath, _ := cmd.Flags().GetString("dbpath")
		if dbpath == "" {
			cobra.CheckErr(msgErrNoDBPath)
		}

		if _, err := os.Stat(dbpath); errors.Is(err, os.ErrNotExist) {
			cobra.CheckErr("local DB does not exists, run 'init' command to create DB")
		}

		store, err := storage.NewSQLiteDB(dbpath)
		if err != nil {
			msg = fmt.Sprint(msgErrInitGRPC, err)
			cobra.CheckErr(msg)
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			msg = fmt.Sprintf("failed initializing GRPC client: %v ", err)
			cobra.CheckErr(msg)
		}

		secretsRemote, err := svc.ListSecrets(token)
		if err != nil {
			msg = fmt.Sprintf("failed to get list of secrets from server: %v", err)
			cobra.CheckErr(msg)
		}

		if len(secretsRemote) != 0 {
			for _, secret := range secretsRemote {
				secretsMapRemote[secret.Name] = secret.Version
			}
		}

		secretsLocal, err := store.SecretList()
		if err != nil {
			msg = fmt.Sprintf("failed to get list of secrets from local DB: %v", err)
			cobra.CheckErr(msg)
		}

		if len(secretsLocal) != 0 {
			// building map from local secrets
			for _, secret := range secretsLocal {
				secretsMapLocal[secret.Name] = secret.Version
			}
		}

		for name, remoteVersion := range secretsMapRemote {
			if localVersion, ok := secretsMapLocal[name]; ok {
				if localVersion < remoteVersion {
					secret, err := svc.GetSecret(token, key, name)
					if err != nil {
						msg = fmt.Sprintf("failed to get secret: %v ", err)
						cobra.CheckErr(msg)
					}
					// updating local secret
					err = store.SecretUpdate(secret)
					if err != nil {
						msg = fmt.Sprintf("failed to update secret: %v", err)
						cobra.CheckErr(msg)
					}
				}
			} else {
				secret, err := svc.GetSecret(token, key, name)
				if err != nil {
					msg = fmt.Sprintf("failed to get secret: %v ", err)
					cobra.CheckErr(msg)
				}
				// Inserting secret into local DB
				err = store.SecretAdd(secret)
				if err != nil {
					msg = fmt.Sprintf("failed to add secret into local DB: %v", err)
					cobra.CheckErr(msg)
				}
			}
		}
		// removing non-existing secrets from local DB
		for name := range secretsMapLocal {
			if _, ok := secretsMapRemote[name]; !ok {
				err = store.SecretDelete(name)
				if err != nil {
					msg = fmt.Sprintf("failed to delete secret in local DB: %v", err)
					cobra.CheckErr(msg)
				}
			}
		}
		fmt.Println("successfully synchronised secret db.")
	},
}
