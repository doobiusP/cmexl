package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Start local cmexl metadata store",
	Long:  `Create the necessary cmexl config files in the current working directory`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
