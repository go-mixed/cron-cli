package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ShellCommand []string

type job struct {
	Name          string         `json:"-" yaml:"-"`
	Schedule      string         `json:"schedule" yaml:"schedule"`
	WorkDirectory string         `json:"work_directory" yaml:"work_directory"`
	Commands      []ShellCommand `json:"commands" yaml:"commands"`
	Env           []string       `json:"env" yaml:"env"`
	Timeout       int64          `json:"timeout" yaml:"timeout"`
	RunningMode   string         `json:"running_mode"` // skip, delay, on-time(default) if last job is running

	StdoutLog string `json:"stdout_log" yaml:"stdout_log"`
	StderrLog string `json:"stderr_log" yaml:"stderr_log"`
	logger    *logger

	id   cron.EntryID
	task *Task
}

func (s ShellCommand) Command() string {
	if len(s) < 1 {
		return ""
	}
	return s[0]
}

func (s ShellCommand) Arguments() []string {
	if len(s) < 2 {
		return nil
	}
	return s[1:]
}

func (s ShellCommand) String() string {
	return strings.Join(s, " ")
}

func (job *job) Run() {
	ctx := job.task.ctx
	// deadline if job.Timeout is valid.
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(job.task.ctx, time.Duration(job.Timeout)*time.Millisecond)
		defer cancel()
	}
for1:
	for _, command := range job.Commands {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				job.logger.Info(fmt.Sprintf("schedule: [%s]. Execution timeout, Skip next command of [%s]\n", job.Schedule, command.String()), "name", job.Name)
			} else if errors.Is(ctx.Err(), context.Canceled) {
				job.logger.Info(fmt.Sprintf("schedule: [%s]. Task exiting, Skip next command of [%s]\n", job.Schedule, command.String()), "name", job.Name)
			}
			break for1
		default:
		}

		job.logger.Info(fmt.Sprintf("schedule: [%s] executing: [%s]\n", job.Schedule, command.String()), "name", job.Name)

		cmd := exec.CommandContext(ctx, command.Command(), command.Arguments()...)
		cmd.Dir = job.WorkDirectory
		cmd.Env = append(os.Environ(), job.Env...)
		cmd.Stdout = job.logger.stdout("name", job.Name)
		cmd.Stderr = job.logger.stderr("name", job.Name)

		if err := cmd.Run(); err != nil {
			job.logger.Error(err, "command execution fail", "name", job.Name)
		}
	}
}

func (job *job) makeLogger(defaultLogger *logger) (err error) {
	job.logger, err = newLogger(job.StdoutLog, job.StderrLog)

	if (job.StdoutLog == "" && job.StderrLog == "") || err != nil {
		job.logger = defaultLogger
	}
	return
}
