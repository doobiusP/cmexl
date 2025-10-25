/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/spf13/cobra"
)


func getCmakePresetsE(cmd *cobra.Command, args []string) error{
	getPresetsCmd := exec.Command("cmake", "--list-presets")
	getPresetsCmd.Stderr = os.Stderr
	namesOnly, namesOnlyPError := cmd.Flags().GetBool("names-only")
	if namesOnlyPError !=nil {
		return fmt.Errorf("%v", namesOnlyPError)
	}
	presetsList, err := getPresetsCmd.Output();
	if len(presetsList)==0 {
		return fmt.Errorf("no presets found")
	}
	if err!= nil {
		return fmt.Errorf("%v", err)
	}
	if namesOnly {
		nameRe := regexp.MustCompile(`"([^"]+)"`)
		presetNames := nameRe.FindAllStringSubmatch(string(presetsList), -1)
		fmt.Printf("Matched: %v", presetNames)
		var namesList string
		for _, name := range presetNames {
			namesList += name[1] + "\n"
		}
		fmt.Printf("%s", namesList)
	} else {
		fmt.Printf("%s", presetsList)
	}
	return nil
}

var presetsCmd = &cobra.Command{
	Use: "presets",
	Short: "List all detected cmake presets in the current working directory",
	Aliases: []string{"lp"},
	RunE: getCmakePresetsE,
}

func init() {
	rootCmd.AddCommand(presetsCmd)
	presetsCmd.Flags().BoolP("names-only", "n", false, "Print only the preset names without quotations and descriptions")

}