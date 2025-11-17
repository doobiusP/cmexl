package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cmexlVer = "0.3.0.2"

var rootCmd = &cobra.Command{
	Use:     "cmexl [command]",
	Short:   "Project bootstrapper & parallel build runner for CMake/C++ projects",
	Version: cmexlVer,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
