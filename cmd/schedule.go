package cmd

import (
	cmutils "cmexl/pkg"
	"errors"

	"github.com/spf13/cobra"
)

var prTypeF = cmutils.ConfigureStr

func execPresetsE(cmd *cobra.Command, args []string) error {
	prType, err := cmutils.MapPresetToType(prTypeF)
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("no arguments provided")
	}

	var prList []cmutils.PresetInfoKey
	for _, arg := range args {
		prKey := cmutils.PresetInfoKey{Name: arg, Type: prType}
		prList = append(prList, prKey)
	}

	err = cmutils.ScheduleCmakePresets(prType, prList)
	return err
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule <preset-name> | subcommand",
	Short: "Schedule preset(s) for execution",
	Long:  `Schedule preset(s) for execution according to configuration rules set in cmexlconf.json`,
	RunE:  execPresetsE,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().VarP(&prTypeF, "ptype", "t", "Type of preset being passed in. Should be one of the cmake preset types.")
}
