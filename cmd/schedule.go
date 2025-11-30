package cmd

import (
	cmutils "cmexl/pkg"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var flags cmutils.ScheduleFlags

func execPresetsE(cmd *cobra.Command, args []string) error {
	prTypeFlag, typeError := cmd.Flags().GetString("type")
	if typeError != nil {
		return typeError
	}
	prType, err := cmutils.MapPresetStrToType(prTypeFlag)
	if err != nil {
		return err
	}

	prMap, prErr := cmutils.GetCmakePresets(prType)
	if prErr != nil {
		return prErr
	}

	if len(args) < 1 {
		return errors.New("no arguments provided")
	}
	var prList []cmutils.PresetInfoKey
	for _, arg := range args {
		prKey := cmutils.PresetInfoKey{Name: arg, Type: prType}
		if _, ok := prMap[prKey]; !ok {
			return fmt.Errorf("%s does not correspond to preset type %s", arg, prType.String())
		}
		prList = append(prList, prKey)
	}

	err = cmutils.ScheduleCmakePresets(prType, prList, prMap, flags)
	return err
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule -t <preset-type> <preset-names>",
	Short: "Schedule preset(s) for execution",
	RunE:  execPresetsE,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().StringP("type", "t", "", "Type of preset being passed in. Should be one of the cmake preset types.")
	flags.SaveEvents = scheduleCmd.Flags().Bool("save-events", false, "Save events picked up by cmexl under .cmexl/events/{presetName}.log")
	flags.Refresh = scheduleCmd.Flags().Bool("refresh", false, "Rebuild binary directory from scratch")
}
