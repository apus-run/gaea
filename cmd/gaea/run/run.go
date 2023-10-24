package run

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run project",
	Long:  "Run project. Example: gaea run",
	Run:   Run,
}

func Run(cmd *cobra.Command, args []string) {

}
