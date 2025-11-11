package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
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
	/*
		Flags:
			--no-vcpkg
			--template {<read-from-templates-dir>}
			--name <name-of-project>
			--add-support [mingw64, clang]
			--add-platform [linux, windows]
			--version <ver>
			--git-init
			--configs {...}
	*/
}
