package secret

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete secret",
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

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			log.Fatal(msgErrInitGRPC, err)
		}

		name, _ := cmd.Flags().GetString(secretName)

		err := svc.DeleteSecret(token, name)
		if err != nil {
			fmt.Println("error getting secret: ", err)
		}
	},
}

func init() {
	DeleteCmd.Flags().StringP(secretName, "n", "", "Secret name.")
	if err := DeleteCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
}
