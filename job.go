package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ShellCommand []string

type job struct {
	Schedule      string         `json:"schedule" yaml:"schedule"`
	WorkDirectory string         `json:"work_directory" yaml:"work_directory"`
	Commands      []ShellCommand `json:"commands" yaml:"commands"`
	Env           []string       `json:"env" yaml:"env"`
	Timeout       int64          `json:"timeout" yaml:"timeout"`
	RunningMode   string         `json:"running_mode"` // skip, delay, on-time(default) if last job is running

	StdoutLog string `json:"stdout_log" yaml:"stdout_log"`
	StderrLog string `json:"stderr_log" yaml:"stderr_log"`
	logger    *logger

	id cron.EntryID
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

func (job *job) Run() {
	ctx := context.Background()
	var cancel context.CancelFunc

	// deadline if job.Timeout is valid.
	if job.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(job.Timeout)*time.Millisecond)
		defer cancel()
	}

	for _, command := range job.Commands {
		job.logger.Infof("[%d]schedule: [%s] executing: %s %s\n", job.id, job.Schedule, command.Command(), strings.Join(command.Arguments(), " "))

		cmd := exec.CommandContext(ctx, command.Command(), command.Arguments()...)
		cmd.Dir = job.WorkDirectory
		cmd.Env = append(os.Environ(), job.Env...)
		cmd.Stdout = job.logger.stdout()
		cmd.Stderr = job.logger.stderr()

		if err := cmd.Run(); err != nil {
			job.logger.Errorf("[%d]fatal: %s\n", job.id, err.Error())
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
