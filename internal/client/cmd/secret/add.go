package secret

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
	"github.com/vkupriya/gophkeeper/internal/client/models"
)

const SecretName string = "name"
const TokenJWT string = "token"

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add or Update secret",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		token := viper.GetViper().GetString(TokenJWT)
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
		var data []byte
		var err error

		name, _ := cmd.Flags().GetString(SecretName)
		meta, _ := cmd.Flags().GetString("metadata")
		stype, _ := cmd.Flags().GetString("stype")
		update, _ := cmd.Flags().GetBool("update")
		file, _ := cmd.Flags().GetString("file")

		if file != "" {
			data, err = os.ReadFile(file)
			if err != nil {
				log.Fatalf("failed to read file %s: %v", file, err)
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
	AddCmd.Flags().StringP(SecretName, "n", "", "Unique secret name.")
	AddCmd.Flags().StringP("data", "d", "", "Secret data.")
	AddCmd.Flags().StringP("metadata", "m", "", "JSON string with secret metadata.")
	AddCmd.Flags().StringP("stype", "t", "text", "Secret type: permitted [text, binary].")
	AddCmd.Flags().StringP("file", "f", "", "File with secret data.")
	AddCmd.Flags().BoolP("update", "u", false, "Update existing secret.")
	AddCmd.MarkFlagsMutuallyExclusive("data", "file")
}
