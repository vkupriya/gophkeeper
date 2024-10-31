package login

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
	"github.com/vkupriya/gophkeeper/internal/client/helpers"
)

// LoginCmd represents the login command.
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Gophkeeper Server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Check host flag
		server, _ := cmd.Flags().GetString("server")
		if server == "" {
			server = viper.GetViper().GetString("server")
		}
		if server == "" {
			log.Fatal("server not specified.")
		} else {
			fmt.Println("server:", server)
		}
		// Check user flag
		user, _ := cmd.Flags().GetString("user")
		if user == "" {
			// Try and get token from config
			token := viper.GetViper().GetString("token")
			if token == "" {
				fmt.Println("Login failed, no user specified.")
			}
		} else {
			// Get password for the user
			password := helpers.GetPassword()

			svc := grpcclient.NewService()

			if err := grpcclient.NewGRPCClient(svc); err != nil {
				log.Fatal("error initializing GRPC client: ", err)
			}
			var token string
			var err error
			reg, _ := cmd.Flags().GetBool("register")
			if reg {
				token, err = svc.Register(user, password)
				if err != nil {
					log.Fatal("error registering user: ", err)
				}
			} else {
				token, err = svc.Login(user, password)
				if err != nil {
					log.Fatal("login error: ", err)
				}
			}
			token = strings.Split(token, ":")[1]
			token = strings.ReplaceAll(token, `"`, "")
			viper.Set("host", server)
			viper.Set("token", token)
			if err = viper.WriteConfig(); err != nil {
				log.Fatal("Error writing configuration file: ", err)
			}
		}
	},
}

func init() {
	LoginCmd.Flags().StringP("server", "s", "127.0.0.1:3200", "GophKeeper Server GRPC.")
	LoginCmd.Flags().StringP("user", "u", "", "Username on GophKeeper Server.")
	LoginCmd.Flags().BoolP("register", "r", false, "Flag to Register User with GophKeeper Server.")
}
