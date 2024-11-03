package login

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcclient "github.com/vkupriya/gophkeeper/internal/client/grpc"
	"github.com/vkupriya/gophkeeper/internal/client/helpers"
)

const serverStr = "server"

// LoginCmd represents the login command.
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Gophkeeper Server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Check host flag
		var msg string
		server, _ := cmd.Flags().GetString(serverStr)
		if server == "" {
			server = viper.GetViper().GetString(serverStr)
		}
		if server == "" {
			cobra.CheckErr("server not specified.")
		}
		// Check user flag
		user, _ := cmd.Flags().GetString("user")
		if user == "" {
			// Try and get token from config
			token := viper.GetViper().GetString("token")
			if token == "" {
				cobra.CheckErr("Login failed, no user specified.")
			}
		} else {
			// Get password for the user
			password := helpers.GetPassword()

			svc := grpcclient.NewService()

			if err := grpcclient.NewGRPCClient(svc, server); err != nil {
				msg = fmt.Sprintf("error initializing GRPC client: %v", err)
				cobra.CheckErr(msg)
			}
			var token string
			var err error
			reg, _ := cmd.Flags().GetBool("register")
			if reg {
				token, err = svc.Register(user, password)
				if err != nil {
					msg = fmt.Sprintf("error registering user: %v", err)
					cobra.CheckErr(msg)
				}
			} else {
				token, err = svc.Login(user, password)
				if err != nil {
					msg = fmt.Sprintf("login error: %v ", err)
					cobra.CheckErr(msg)
				}
			}
			viper.Set(serverStr, server)
			viper.Set("token", token)
			if err = viper.WriteConfig(); err != nil {
				msg = fmt.Sprintf("Error writing configuration file: %v", err)
				cobra.CheckErr(msg)
			}
		}
	},
}

func init() {
	LoginCmd.Flags().StringP("server", "s", "127.0.0.1:3200", "GophKeeper Server GRPC.")
	LoginCmd.Flags().StringP("user", "u", "", "Username on GophKeeper Server.")
	LoginCmd.Flags().BoolP("register", "r", false, "Flag to Register User with GophKeeper Server.")
}
