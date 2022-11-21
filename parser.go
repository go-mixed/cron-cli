package main

import (
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func (t *Task) parseArguments(args []string) ([]*job, error) {
	if len(args) <= 0 {
		return nil, nil
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

	var jobs []*job
	for _, line := range jobLines {
		if len(line) <= 1 {
			panic("invalid schedule and command: " + strings.Join(line, " "))
		}
		jobs = append(jobs, &job{
			Schedule: line[0],
			Command:  strings.Join(line[1:], " "),
		})

	}
	return jobs, nil
}

func (t *Task) parseCronFile(filePath string) ([]*job, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(scanLines)

	var jobs []*job
	for scanner.Scan() {
		line := scanner.Text()

		segments := strings.SplitN(line, " ", 6)
		if len(segments) != 6 {
			return nil, fmt.Errorf("schedule must be 5 fields in \"%s\", like [* * * * *], yours: %s", filePath, line)
		}

		jobs = append(jobs, &job{Schedule: strings.Join(segments[0:5], " "), Command: segments[5]})
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		raw := dropCR(data[0:i])
		// A backslash escapes the next character from being interpreted by the shell.
		// If the next character after the backslash is a newline character,
		// then that newline will not be interpreted as the end of the command by the shell.
		// Instead, it effectively allows a command to span multiple lines.
		if len(raw) >= 1 && raw[len(raw)-1] == '\\' {
			return 0, nil, nil
		}
		// We have a full newline-terminated line.
		return i + 1, raw, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func (t *Task) parseYamlFile(filePath string) ([]*job, error) {
	type yamlConfig struct {
		Schedules []*job `json:"schedules" yaml:"schedules"`
	}
	var actual yamlConfig
	if content, err := os.ReadFile(filePath); err != nil {
		return nil, fmt.Errorf("read yaml file error: %w", err)
	} else {
		if err = yaml.Unmarshal(content, &actual); err != nil {
			return nil, fmt.Errorf("unmarshal yaml file \"%s\" error: %w", filePath, err)
		}
	}
	return actual.Schedules, nil
}
