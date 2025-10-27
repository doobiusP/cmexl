/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	cmutils "cmexl/pkg"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

type PresetStr_t string

const (
	ConfigureStr PresetStr_t = "configure"
	BuildStr     PresetStr_t = "build"
	TestStr      PresetStr_t = "test"
	PackageStr   PresetStr_t = "package"
	WorkflowStr  PresetStr_t = "workflow"
)

func (prS *PresetStr_t) String() string {
	return string(*prS)
}

func (preS *PresetStr_t) Set(v string) error {
	switch v {
	case "configure", "build", "test", "package", "workflow":
		*preS = PresetStr_t(v)
		return nil
	default:
		return errors.New(`must be one of {"configure", "build", "test", "package", "workflow"}`)
	}
}

func (prS *PresetStr_t) Type() string {
	return "presetStr_t"
}

var presetTypeFlag = ConfigureStr

func execCMakePresetE(cmd *cobra.Command, args []string) error {
	presetType, err := cmutils.MapPresetToType(presetTypeFlag.String())
	if err != nil {
		return err
	}
	prMap, prErr := cmutils.GetCmakePresets(presetType)
	if prErr != nil {
		return prErr
	}

	if len(args) != 1 {
		return errors.New("only 1 argument should be passed")
	}
	var cmakeArgs []string
	var cmakeCmd string

	prKey := cmutils.PresetInfoKey{Name: args[0], Type: presetType}
	if _, ok := prMap[prKey]; !ok {
		return errors.New("cant find the preset")
	}
	cmakeCmd = "cmake"
	switch presetType {
	case cmutils.Configure:
		break
	case cmutils.Build:
		cmakeArgs = append(cmakeArgs, "--build")
	case cmutils.Workflow:
		cmakeArgs = append(cmakeArgs, "--workflow")
	case cmutils.Test:
		cmakeCmd = "ctest"
	case cmutils.Package:
		cmakeCmd = "cpack"
	default:
		return errors.New("got unexpected Preset_t type")
	}
	cmakeArgs = append(cmakeArgs, "--preset")
	cmakeArgs = append(cmakeArgs, args[0])

	fmt.Printf("%s %s\n", cmakeCmd, cmakeArgs)
	cmakeExecCmd := exec.Command(cmakeCmd, cmakeArgs...)
	cmakeExecCmd.Stdout = os.Stdout
	cmakeExecCmd.Stderr = os.Stderr

	start := time.Now()
	cmakeErr := cmakeExecCmd.Run()
	if cmakeErr != nil {
		return fmt.Errorf("cmake execution error: %w", cmakeErr)
	}
	elapsed := time.Since(start) // Calculate the elapsed time
	fmt.Printf("Execution took %s\n", elapsed)
	return nil
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule <preset-name> | subcommand",
	Short: "Schedule preset(s) for execution",
	Long:  `Schedule preset(s) for execution according to configuration rules set in cmexlconf.json`,
	RunE:  execCMakePresetE,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().VarP(&presetTypeFlag, "ptype", "t", "Type of preset being passed in. Should be one of the cmake preset types.")
}
