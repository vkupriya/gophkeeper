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

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get secret",
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
			log.Fatal("Missing secretkey, update configuration file")
		}
		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			log.Fatal(msgErrInitGRPC, err)
		}

		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			log.Fatal("no secret provided.")
		}

		secret, err := svc.GetSecret(token, key, name)
		if err != nil {
			if errors.Is(err, grpcclient.ErrServerUnavailable) {
				fmt.Println("server is unavailable, attempting to read secret from local DB.")
				dbpath, _ := cmd.Flags().GetString("dbpath")
				if dbpath == "" {
					log.Fatal(msgErrNoDBPath)
				}
				store, err := storage.NewSQLiteDB(dbpath)
				if err != nil {
					log.Fatal("Error in setting up DB: ", err)
				}
				defer func() {
					if err := store.DB.Close(); err != nil {
						log.Fatal("failed to close local DB")
					}
				}()

				secret, err = store.SecretGet(name)
				if err != nil {
					if errors.Is(err, storage.ErrSecretNotFound) {
						log.Fatal("secret not found in local DB")
					}
					fmt.Println("failed to read secrets from local DB", err)
				}
			} else {
				log.Fatal("error getting secret: ", err)
			}
		}

		if secret != nil {
			switch secret.Type {
			case "text":
				res, _ := json.MarshalIndent(models.SecretPrint{
					Name:    secret.Name,
					Type:    secret.Type,
					Meta:    secret.Meta,
					Data:    string(secret.Data),
					Version: secret.Version,
				}, "", "    ")
				fmt.Println(string(res))
			case "card":
				res, _ := json.MarshalIndent(models.SecretPrint{
					Name:    secret.Name,
					Type:    secret.Type,
					Meta:    secret.Meta,
					Data:    string(secret.Data),
					Version: secret.Version,
				}, "", "    ")
				fmt.Println(string(res))
			default:
				res, _ := json.MarshalIndent(secret, "", "    ")
				fmt.Println(string(res))
			}
		}

	},
}

func init() {
	GetCmd.Flags().StringP(secretName, "n", "", "Secret name.")
	if err := GetCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
}
