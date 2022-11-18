package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
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
			docker, _ := cmd.PersistentFlags().GetBool("docker")
			task := buildTask(args, logPath)
			task.isDocker = docker

			if cmd.PersistentFlags().Changed("config") {
				configs, _ := cmd.PersistentFlags().GetStringSlice("config")
				if err := task.LoadSettings(configs...); err != nil {
					panic(err.Error())
				}
			}

			task.Start()
			task.ListenStopSignal()
			task.Wait()
		},
	}

	rootCmd.PersistentFlags().Bool("docker", false, "run this application in docker")
	rootCmd.PersistentFlags().StringSliceP("config", "c", []string{}, "the path of config files or directories")
	rootCmd.PersistentFlags().StringP("log", "l", "", "the path of log file")

	err := rootCmd.Execute()
	if err != nil {
		panic(err.Error())
	}
}

func buildTask(args []string, logPath string) *Task {
	log, err := newLogger(logPath, "")
	if err != nil {
		panic("create logger error: " + err.Error())
	}
	var task = NewTask(log)

	if len(args) <= 0 {
		return task
	}

	// 通过 -- 分割
	var jobLines [][]string
	var lastSeparator int
	for i, arg := range args {
		if arg == "--" {
			jobLines = append(jobLines, args[lastSeparator:i])
			lastSeparator = i + 1
		} else if i == len(args)-1 {
			jobLines = append(jobLines, args[lastSeparator:])
		}
	}

	for i, line := range jobLines {
		if len(line) <= 1 {
			panic("invalid schedule and command: " + strings.Join(line, " "))
		}
		if err = task.AddJob(&job{
			Name:     fmt.Sprintf("argument-%d", i),
			Schedule: line[0],
			Commands: []ShellCommand{line[1:]},
		}); err != nil {
			panic(err.Error())
		}
	}

	return task
}
