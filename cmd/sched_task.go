package cmd

import (
	cmutils "cmexl/pkg"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func schedGroupE(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("can only accept 1 task")
	}
	var C cmutils.Config
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
	PersistentPreRunE: cmutils.FindCmexlConf,
	RunE:              schedGroupE,
	Args:              cobra.ExactArgs(1),
}

func init() {
	scheduleCmd.AddCommand(schedTaskCmd)
}
