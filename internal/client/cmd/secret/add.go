package secret

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
	"github.com/vkupriya/gophkeeper/internal/client/models"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add or Update secret",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
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
		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc, server); err != nil {
			cobra.CheckErr(err)
		}
		var data []byte
		var err error

		name, _ := cmd.Flags().GetString(secretName)
		meta, _ := cmd.Flags().GetString("metadata")
		stype, _ := cmd.Flags().GetString("stype")
		update, _ := cmd.Flags().GetBool("update")
		file, _ := cmd.Flags().GetString("file")

		if file != "" {
			data, err = os.ReadFile(file)
			if err != nil {
				msg := fmt.Sprintf("failed to read file %s: %v", file, err)
				cobra.CheckErr(msg)
			}
		} else {
			dataStr, _ := cmd.Flags().GetString("data")
			data = []byte(dataStr)
		}

		secret := &models.Secret{
			Name: name,
			Data: data,
			Type: stype,
			Meta: meta,
		}
		if update {
			err := svc.UpdateSecret(token, key, secret)
			if err != nil {
				fmt.Println("error updating secret: ", err)
			}
		} else {
			err := svc.AddSecret(token, key, secret)
			if err != nil {
				fmt.Println("error adding secret: ", err)
			}
		}
	},
}

func init() {
	AddCmd.Flags().StringP(secretName, "n", "", "Unique secret name.")
	AddCmd.Flags().StringP("data", "d", "", "Secret data.")
	AddCmd.Flags().StringP("metadata", "m", "", "JSON string with secret metadata.")
	AddCmd.Flags().StringP("stype", "t", "text", "Secret type: permitted [text, binary, card].")
	AddCmd.Flags().StringP("file", "f", "", "File with secret data.")
	AddCmd.Flags().BoolP("update", "u", false, "Update existing secret.")
	AddCmd.MarkFlagsMutuallyExclusive("data", "file")
}
