/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	cmutils "cmexl/pkg"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func printPresets(cmd *cobra.Command, prType cmutils.Preset_t) error {
	prMap, presetErr := cmutils.GetCmakePresets(prType)
	if presetErr != nil {
		return presetErr
	}
	namesOnly, namesOnlyPError := cmd.PersistentFlags().GetBool("names-only")
	if namesOnlyPError != nil {
		return fmt.Errorf("%v", namesOnlyPError)
	}
	if namesOnly {
		for prKey := range prMap {
			prKeyTypeStr, _ := prKey.Type.String()
			fmt.Printf("(%s, %s)\n", prKey.Name, prKeyTypeStr)
		}
	} else {
		for _, prVal := range prMap {
			fmt.Printf("(%s)\n", prVal.String())
		}
	}
	return nil
}

func getCmakePresetsE(cmd *cobra.Command, args []string) error {
	allArgFound := false
	for _, arg := range args {
		if cmutils.PresetIsAllowed(arg) {
			if arg == "all" {
				allArgFound = true
			}
		} else {
			return errors.New("found unexpected preset type in list command")
		}
	}

	if allArgFound {
		if err := printPresets(cmd, cmutils.All); err != nil {
			return err
		}
	} else {
		for _, arg := range args {
			prType, prErr := cmutils.MapPresetToType(arg)
			if prErr != nil {
				return prErr
			}
			if err := printPresets(cmd, prType); err != nil {
				return err
			}
		}
	}
	return nil
}

var listCmd = &cobra.Command{
	Use:   "list {configure, build, test, package, workflow} [-n]",
	Short: "List the different cmake presets detected",
	Long: `List the cmake presets detected in the current working directory
according to a category from {configure, build, test, package, workflow}.
By default, prints all presets found (equivalent list as cmake --list-presets)`,
	RunE: getCmakePresetsE,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolP("names-only", "n", false, "Print only the preset names without quotations and descriptions")
}
