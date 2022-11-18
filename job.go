package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type job struct {
	Name          string   `json:"-" yaml:"-"`
	Schedule      string   `json:"schedule" yaml:"schedule"`
	WorkDirectory string   `json:"work_directory" yaml:"work_directory"`
	Commands      []string `json:"commands" yaml:"commands"`
	Env           []string `json:"env" yaml:"env"`
	Timeout       int64    `json:"timeout" yaml:"timeout"`
	RunningMode   string   `json:"running_mode"` // skip, delay, on-time(default) if last job is running

	StdoutLog string `json:"stdout_log" yaml:"stdout_log"`
	StderrLog string `json:"stderr_log" yaml:"stderr_log"`
	logger    *logger

	id   cron.EntryID
	task *Task
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
				job.logger.Info(fmt.Sprintf("execution timeout, Skip next command of [%s]\n", command), "schedule", job.Schedule, "name", job.Name)
			} else if errors.Is(ctx.Err(), context.Canceled) {
				job.logger.Info(fmt.Sprintf("task exiting, Skip next command of [%s]\n", command), "schedule", job.Schedule, "name", job.Name)
			}
			break for1
		default:
		}

		var actualCommand []string
		if job.task.isDocker {
			actualCommand = append([]string{"nsenter", "-t", "1", "-m", "-u", "-n", "-i", "sh", "-c"}, command)
		} else if runtime.GOOS == "windows" {
			actualCommand = append([]string{"c:\\windows\\system32\\cmd.exe", "/C"}, command)
		} else {
			actualCommand = append([]string{"sh", "-c"}, command)
		}

		job.logger.Info("executing", "schedule", job.Schedule, "command", command, "name", job.Name)

		cmd := exec.CommandContext(ctx, actualCommand[0], actualCommand[1:]...)
		cmd.Dir = job.WorkDirectory
		cmd.Env = append(os.Environ(), job.Env...)
		cmd.Stdout = job.logger.stdout("name", job.Name, "command", command)
		cmd.Stderr = job.logger.stderr("name", job.Name, "command", command)

		if err := cmd.Run(); err != nil {
			job.logger.Error(err, "command execution fail", "schedule", job.Schedule, "command", command, "name", job.Name)
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
