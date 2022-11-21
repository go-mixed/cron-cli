package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type job struct {
	Schedule      string   `json:"schedule" yaml:"schedule"`
	WorkDirectory string   `json:"work_directory" yaml:"work_directory"` // disabled in docker mode
	Command       string   `json:"command" yaml:"command"`
	Env           []string `json:"env" yaml:"env"`
	Timeout       int64    `json:"timeout" yaml:"timeout"`
	RunningMode   string   `json:"running_mode"` // [skip, delay, on-time(default)] if last job is running

	StdoutLog string `json:"stdout_log" yaml:"stdout_log"`
	StderrLog string `json:"stderr_log" yaml:"stderr_log"`
	logger    *logger

	id        cron.EntryID
	task      *Task
	shellFile string
}

func (job *job) saveShellFile() {
	if job.shellFile != "" {
		path := job.shellFile
		if job.task.InDocker() {
			path = filepath.Join(job.task.rootPathInDocker, job.shellFile)
		}
		_ = os.WriteFile(path, []byte(job.Command), 0x644)
	}
}

func (job *job) deleteShellFile() {
	if job.shellFile != "" {
		path := job.shellFile
		if job.task.InDocker() {
			path = filepath.Join(job.task.rootPathInDocker, job.shellFile)
		}

		_ = os.Remove(path)
	}
}

func (job *job) Run() {
	ctx := job.task.ctx
	// deadline if job.Timeout is valid.
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(job.task.ctx, time.Duration(job.Timeout)*time.Millisecond)
		defer cancel()
	}

	var actualCommand []string
	if job.task.InDocker() {
		actualCommand = []string{"nsenter", "-t", "1", "-m", "-u", "-n", "-i", "/usr/bin/sh", job.shellFile}
	} else if runtime.GOOS == "windows" {
		actualCommand = []string{"c:\\windows\\system32\\cmd.exe", job.shellFile}
	} else {
		actualCommand = []string{"/usr/bin/sh", job.shellFile}
	}

	var truncatedCmd = truncateText(job.Command, 40)
	job.logger.Info("executing", "schedule", job.Schedule, "command", truncatedCmd, "id", job.id)

	cmd := exec.CommandContext(ctx, actualCommand[0], actualCommand[1:]...)
	if !job.task.InDocker() {
		cmd.Dir = job.WorkDirectory
	}
	cmd.Env = append(os.Environ(), job.Env...)
	cmd.Stdout = job.logger.stdout("command", truncatedCmd, "id", job.id)
	cmd.Stderr = job.logger.stderr("command", truncatedCmd, "id", job.id)

	if err := cmd.Run(); err != nil {
		job.logger.Error(err, "command execution fail", "schedule", job.Schedule, "command", truncatedCmd, "id", job.id)
	}

}

func (job *job) makeLogger(defaultLogger *logger) (err error) {
	job.logger, err = newLogger(job.StdoutLog, job.StderrLog)

	if (job.StdoutLog == "" && job.StderrLog == "") || err != nil {
		job.logger = defaultLogger
	}
	return
}
