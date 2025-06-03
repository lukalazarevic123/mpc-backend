package cmd

import (
	"mpc-backend/server"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const hostFlag = "host"
const portFlag = "port"

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs HTTP sever",
	Run:   run,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String(hostFlag, "localhost", "Host on which to run the server")
	serveCmd.Flags().IntP(portFlag, "p", 8080, "Port on which to run the server")

	// Bind the flags to the configuration
	_ = viper.BindPFlag("server.host", serveCmd.Flags().Lookup(hostFlag))
	_ = viper.BindPFlag("server.port", rootCmd.Flags().Lookup(portFlag))
}

func run(_ *cobra.Command, _ []string) {
	err := server.NewServer(configuration)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
