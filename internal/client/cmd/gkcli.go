package gkcli

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vkupriya/gophkeeper/internal/client/cmd/login"
	"github.com/vkupriya/gophkeeper/internal/client/cmd/secret"
)

var cfgFile string
var server string
var dbpath string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gkcli",
	Short: "gkcli - GophKeeper CLI client",
	Long:  `gkcli is a CLI for GophKeeper secret manager.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gk.yaml)")
	rootCmd.PersistentFlags().StringVar(&server, "server", "127.0.0.1:3200", "gophkeeper server address:port")
	rootCmd.PersistentFlags().StringVar(&dbpath, "db", "./secrets.db", "path to sqlite local database")
	rootCmd.AddCommand(InitCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(login.LoginCmd)
	rootCmd.AddCommand(secret.SecretCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "something went wrong: %s", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gk.yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		_, err = os.Create(home + "/.gk.yaml")
		if err != nil {
			log.Fatal()
		}
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())
}
