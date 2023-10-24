package upgarde

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the gaea tools",
	Long:  "Upgrade the gaea tools. Example: gaea upgrade",
	Run:   Run,
}

// Run upgrade the gaea tools.
func Run(cmd *cobra.Command, args []string) {

}
