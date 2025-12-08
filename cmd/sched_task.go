package cmd

import (
	cmutils "cmexl/pkg"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cmexlConfigFileName = "cmexlconf"
	cmexlConfigFileType = "json"
)

type Config struct {
	InitSettings InitSettings `mapstructure:"init_settings"`
	Tasks        []Task       `mapstructure:"tasks"`
}

type InitSettings struct {
	Name     string `mapstructure:"name"`
	Sname    string `mapstructure:"sname"`
	Lname    string `mapstructure:"lname"`
	Uname    string `mapstructure:"uname"`
	UseVcpkg bool   `mapstructure:"use_vcpkg"`
}

type Task struct {
	Name      string   `mapstructure:"name"`
	Workflows []string `mapstructure:"workflows"`
}

func findCmexlConf(cmd *cobra.Command, args []string) error {
	viper.AddConfigPath(".")
	viper.AddConfigPath(".cmexl/")
	viper.SetConfigName(cmexlConfigFileName)
	viper.SetConfigType(cmexlConfigFileType)
	err := viper.ReadInConfig()
	return err
}

func schedGroupE(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("can only accept 1 task")
	}
	var C Config
	if err := viper.Unmarshal(&C); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	for _, task := range C.Tasks {
		if task.Name == args[0] {
			return execPresetsE(cmutils.Workflow, task.Workflows)
		}
	}

	return fmt.Errorf("can't find task called %s", args[0])
}

var schedTaskCmd = &cobra.Command{
	Use:               "task <task-name>",
	Short:             "schedule cmexl task for execution",
	PersistentPreRunE: findCmexlConf,
	RunE:              schedGroupE,
	Args:              cobra.ExactArgs(1),
}

func init() {
	scheduleCmd.AddCommand(schedTaskCmd)
}
