package cmd

import (
	"github.com/dyammarcano/fullcycle_api/components"

	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Client to get data from external API",
	Run: func(cmd *cobra.Command, args []string) {
		components.Client(cmd.Context(), cmd.Flag("port").Value.String())
	},
}

func init() {
	clientCmd.Flags().StringP("port", "p", "8080", "Port to listen")
	rootCmd.AddCommand(clientCmd)
}
