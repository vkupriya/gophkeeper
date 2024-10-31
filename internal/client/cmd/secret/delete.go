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

		token := viper.GetViper().GetString("token")
		if token == "" {
			fmt.Printf("Missing token, please, login.")
		}

		svc := grpcclient.NewService()

		if err := grpcclient.NewGRPCClient(svc); err != nil {
			fmt.Println("error initializing GRPC client: ", err)
		}

		name, _ := cmd.Flags().GetString("name")

		err := svc.DeleteSecret(token, name)
		if err != nil {
			fmt.Println("error getting secret: ", err)
		}
	},
}

func init() {
	DeleteCmd.Flags().StringP("name", "n", "", "Secret name.")
	if err := DeleteCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
}
