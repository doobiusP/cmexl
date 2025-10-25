/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cmexlVer = "0.1.0.0"

var rootCmd = &cobra.Command{
	Use:   "cmexl",
	Short: "Autorunner for CMake presets",
	Long: `Automatically runs and records CMake workflow outputs for your CMake presets:`,
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


