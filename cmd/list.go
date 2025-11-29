package cmd

import (
	cmutils "cmexl/pkg"
	"encoding/json"
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

	var presetDetails []map[string]any

	if namesOnly {
		for prKey := range prMap {
			presetDetails = append(presetDetails, map[string]any{
				"name": prKey.Name,
				"type": prKey.Type.String(),
			})
		}
	} else {
		for _, prVal := range prMap {
			presetDetails = append(presetDetails, map[string]any{
				"name":    prVal.Name,
				"display": prVal.DisplayName,
				"type":    prVal.Type.String(),
				"file":    prVal.File,
			})
		}
	}

	jsonData, err := json.MarshalIndent(presetDetails, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
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

	if allArgFound || len(args) == 0 {
		if err := printPresets(cmd, cmutils.All); err != nil {
			return err
		}
	} else {
		for _, arg := range args {
			prType, prErr := cmutils.MapPresetStrToType(arg)
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
