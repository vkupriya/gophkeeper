package secret

import (
	"github.com/spf13/cobra"
)

const (
	msgErrMissingGRPCServer = "missing grpc server address and port"
	msgErrMissingToken      = "missing user token, please, login"
	msgErrInitGRPC          = "error initializing GRPC client: "
)

const secretName string = "name"
const tokenJWT string = "token"
const hostGRPC string = "server"

var SecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "secret crud commands",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	SecretCmd.AddCommand(AddCmd)
	SecretCmd.AddCommand(ListCmd)
	SecretCmd.AddCommand(GetCmd)
	SecretCmd.AddCommand(DeleteCmd)
	SecretCmd.AddCommand(SyncCmd)
}
