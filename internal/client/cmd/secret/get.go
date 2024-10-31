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

		token := viper.GetViper().GetString("token")
		if token == "" {
			fmt.Printf("Missing token, please, login.")
		}
		key := viper.GetViper().GetString("secretkey")
		if key == "" {
			fmt.Printf("Missing secretkey, update configuration file.")
		}
		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc); err != nil {
			fmt.Println("error initializing GRPC client: ", err)
		}

		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			log.Fatal("no secret provided.")
		}

		secret, err := svc.GetSecret(token, key, name)
		if err != nil {
			if errors.Is(err, grpcclient.ErrServerUnavailable) {
				fmt.Println("server is unavailable, attempting to read secret from local DB.")
				store, err := storage.NewSQLiteDB()
				if err != nil {
					log.Fatal("Error in setting up DB: ", err)
				}
				defer store.DB.Close()
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
	GetCmd.Flags().StringP("name", "n", "", "Secret name.")
	if err := GetCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
}
