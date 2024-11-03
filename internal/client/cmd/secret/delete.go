package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete secret",
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

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			msg = fmt.Sprint(msgErrInitGRPC, err)
			cobra.CheckErr(msg)
		}

		name, _ := cmd.Flags().GetString(secretName)

		err := svc.DeleteSecret(token, name)
		if err != nil {
			cobra.CheckErr(err)
		}
	},
}

func init() {
	DeleteCmd.Flags().StringP(secretName, "n", "", "Secret name.")
	if err := DeleteCmd.MarkFlagRequired("name"); err != nil {
		cobra.CheckErr(err)
	}
}
