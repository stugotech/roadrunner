package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stugotech/goconfig"
	"github.com/stugotech/roadrunner/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Runs a server which solves ACME challenges",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = viper.AddConfigPath
		srv, err := server.NewServer(server.ReadConfig(goconfig.Viper()))
		if err != nil {
			return logger.Errore(err)
		}
		if err = srv.Listen(); err != nil {
			return logger.Errore(err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
