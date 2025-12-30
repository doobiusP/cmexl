package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var taskName string
var presetNamesList []string

type Config struct {
	InitSettings InitSettings `mapstructure:"init_settings"`
	Tasks        []Task       `mapstructure:"tasks"`
}

type InitSettings struct {
	Name     string `mapstructure:"name"`
	Sname    string `mapstructure:"sname"`
	Lname    string `mapstructure:"lname"`
	Uname    string `mapstructure:"uname"`
	Template string `mapstructure:"template"`
}

type Task struct {
	Name      string   `mapstructure:"name"`
	Workflows []string `mapstructure:"workflows"`
}

func execPresets(workflowPresets []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := os.MkdirAll(".cmexl", 0755)
	if err != nil {
		return err
	}
	var failedPresets []string

	numPresets := len(workflowPresets)
	for idx, workflowName := range workflowPresets {
		color.HiBlue("(%d/%d) now running %s", idx+1, numPresets, workflowName)

		logFilePath := fmt.Sprintf(".cmexl/%s.log", workflowName)
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			logFile.Close()
			return fmt.Errorf("can't open %s for write : %v", logFilePath, err)
		}
		logMw := io.MultiWriter(os.Stdout, logFile)

		args := []string{"--workflow", "--preset", workflowName, "--fresh"}
		cmakeCmd := exec.CommandContext(ctx, "cmake", args...)
		cmakeCmd.Stdout = logMw
		cmakeCmd.Stderr = logMw

		err = cmakeCmd.Run()
		logFile.Close()

		if err != nil {
			failedPresets = append(failedPresets, fmt.Sprintf("%s failed : %v, see %s", workflowName, err, logFilePath))
			color.Red("%s failed", workflowName)
		} else {
			color.Green("%s success", workflowName)
		}
	}
	numFailedPresets := len(failedPresets)
	color.Green("\nCMEXL: %d / %d suceeded\n", numPresets-numFailedPresets, numPresets)
	if numFailedPresets > 0 {
		color.Red("Failures:")
		for _, failMsg := range failedPresets {
			fmt.Println(failMsg)
		}
	}
	return nil
}

func execTaskE(cmd *cobra.Command, args []string) error {
	if len(presetNamesList) > 0 {
		return execPresets(presetNamesList)
	}
	if len(taskName) <= 0 {
		return fmt.Errorf("need to specify non-empty task name")
	}

	viper.AddConfigPath(".")
	viper.SetConfigFile("cmexlconf.json")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var cmexlConf Config
	if err := viper.Unmarshal(&cmexlConf); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	for _, confTask := range cmexlConf.Tasks {
		if confTask.Name == taskName {
			return execPresets(confTask.Workflows)
		}
	}
	return fmt.Errorf("did not find task name in cmexlconf.json")
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Batch build a set of cmake workflows (serial)",
	RunE:  execTaskE,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&taskName, "task", "t", "", "name of the build task in cmexlconf.json")
	buildCmd.Flags().StringSliceVarP(&presetNamesList, "presets", "p", []string{}, "manually specified list of workflow presets to execute")
}
