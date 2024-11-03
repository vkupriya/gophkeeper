package secret

import (
	"encoding/json"
	"errors"
	"fmt"

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
		var msg string
		var secrets []*models.SecretItem
		server := viper.GetViper().GetString(hostGRPC)
		if server == "" {
			cobra.CheckErr(msgErrMissingGRPCServer)
		}

		token := viper.GetViper().GetString(tokenJWT)
		if token == "" {
			cobra.CheckErr(msgErrMissingToken)
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			msg = fmt.Sprint(msgErrInitGRPC, err)
			cobra.CheckErr(msg)
		}

		secrets, err := svc.ListSecrets(token)
		if err != nil {
			if errors.Is(err, grpcclient.ErrServerUnavailable) {
				fmt.Println("server is unavailable, attempting to read secrets from local DB.")
				dbpath, _ := cmd.Flags().GetString("dbpath")
				if dbpath == "" {
					cobra.CheckErr(msgErrNoDBPath)
				}
				store, err := storage.NewSQLiteDB(dbpath)
				if err != nil {
					msg = fmt.Sprintf("Error in setting up DB: %v", err)
					cobra.CheckErr(msg)
				}
				defer func() {
					if err := store.DB.Close(); err != nil {
						cobra.CheckErr("failed to close local DB")
					}
				}()

				secrets, err = store.SecretList()
				if err != nil {
					msg = fmt.Sprintf("failed to read secrets from local DB: %v", err)
					cobra.CheckErr(msg)
				}

			} else {
				msg = fmt.Sprintf("error getting list of secrets: %v", err)
				cobra.CheckErr(msg)
			}
		}

		if len(secrets) != 0 {
			res, err := json.MarshalIndent(secrets, "", "   ")
			if err != nil {
				cobra.CheckErr(err)
			}
			fmt.Println(string(res))
		}
	},
}
