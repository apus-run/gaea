package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/apus-run/gaea/cmd/gaea/run"
	"github.com/apus-run/gaea/cmd/gaea/upgarde"
)

var rootCmd = &cobra.Command{
	Use:   "Gaea",
	Short: "Gaea: 基于gRPC业务开发框架",
	Long:  `Gaea: 基于gRPC业务开发框架`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(run.Cmd)
	rootCmd.AddCommand(upgarde.Cmd)
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
