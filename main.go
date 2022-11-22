package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

type cmdOptions struct {
	rootPathInDocker string
	configs          []string
	log              string
	test             bool
}

func main() {

	options := cmdOptions{}

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

			task := buildTask(options.log, options.test)
			task.rootPathInDocker = options.rootPathInDocker

			if err := task.LoadArguments(args); err != nil {
				panic(err.Error())
			}

			if err := task.LoadConfigs(options.configs...); err != nil {
				panic(err.Error())
			}

			task.Start()
			task.ListenStopSignal(func() {
				task.Stop()
			})

			task.Wait()
		},
	}

	rootCmd.PersistentFlags().StringVar(&options.rootPathInDocker, "root-path-in-docker", "/", "What the mounted path of / of host os. Implied meaning: run this application in docker container")
	rootCmd.PersistentFlags().StringSliceVarP(&options.configs, "config", "c", []string{}, "the path of config files or directories")
	rootCmd.PersistentFlags().StringVarP(&options.log, "log", "l", "", "the path of log file")
	rootCmd.PersistentFlags().BoolVar(&options.test, "test", false, "execute all commands immediately and quit")

	err := rootCmd.Execute()
	if err != nil {
		panic(err.Error())
	}
}

func buildTask(logPath string, test bool) *Task {
	log, err := newLogger(logPath, "")
	if err != nil {
		panic("create logger error: " + err.Error())
	}
	task := NewTask(log)
	task.testMode = test

	return task
}
