package cmd

import (
	"github.com/dyammarcano/fullcycle_api/components"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var filePath string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server to get data from external API",
	Run: func(cmd *cobra.Command, args []string) {
		components.Server(cmd.Context(), cmd.Flag("port").Value.String(), filePath)
	},
}

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filePath = filepath.Join(dir, "database.sqlite3")
	serverCmd.Flags().StringP("port", "p", "8080", "Port to listen")
	rootCmd.AddCommand(serverCmd)
}
