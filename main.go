package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "cron -- [schedule1] [command1] [args1...] -- [schedule2] [command2] [args2...]",
		Short: "version: 1.1 \nexample: cron -- \"* * * * * *\" echo 'hello'",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.PersistentFlags().Changed("config") && len(args) < 2 {
				return fmt.Errorf("At least 2 arguments or set --config\n\n")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			logPath, _ := cmd.PersistentFlags().GetString("log")
			rootPathInDocker, _ := cmd.PersistentFlags().GetString("root-path-in-docker")

			task := buildTask(logPath)
			task.rootPathInDocker = rootPathInDocker

			if err := task.LoadArguments(args); err != nil {
				panic(err.Error())
			}

			if cmd.PersistentFlags().Changed("config") {
				configs, _ := cmd.PersistentFlags().GetStringSlice("config")
				if err := task.LoadConfigs(configs...); err != nil {
					panic(err.Error())
				}
			}

			task.Start()
			task.ListenStopSignal()
			task.Wait()
		},
	}

	rootCmd.PersistentFlags().String("root-path-in-docker", "/", "What the mounted path of / of host os. Implied meaning: run this application in docker container")
	rootCmd.PersistentFlags().StringSliceP("config", "c", []string{}, "the path of config files or directories")
	rootCmd.PersistentFlags().StringP("log", "l", "", "the path of log file")

	err := rootCmd.Execute()
	if err != nil {
		panic(err.Error())
	}
}

func buildTask(logPath string) *Task {
	log, err := newLogger(logPath, "")
	if err != nil {
		panic("create logger error: " + err.Error())
	}
	return NewTask(log)
}
