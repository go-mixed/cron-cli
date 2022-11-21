package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (t *Task) LoadArguments(args []string) error {
	jobs, err := t.parseArguments(args)
	if err != nil {
		return err
	}

	if err = t.AddJob("argument", jobs...); err != nil {
		return err
	}

	return nil
}

func (t *Task) LoadConfigs(configs ...string) error {

	var filenames []string
	var err error
	for _, path := range configs {
		if stat, err := os.Stat(path); err != nil {
			return err
		} else if stat.IsDir() {
			if filenames, err = findFiles(path); err != nil {
				return err
			}
		} else {
			if path, err = filepath.Abs(path); err != nil {
				return err
			}
			filenames = append(filenames, path)
		}
	}
	for _, filename := range filenames {
		var jobs []*job
		switch strings.ToLower(filepath.Ext(filename)) {
		case ".yaml", ".yml":
			jobs, err = t.parseYamlFile(filename)
		case ".cron":
			jobs, err = t.parseCronFile(filename)
		default:
			err = fmt.Errorf("the extensions of config file must be [.yaml .yml .cron], yours: %s", filename)
		}

		if err != nil {
			return err
		}

		if err = t.AddJob(filename, jobs...); err != nil {
			return fmt.Errorf("config \"%s\" error: %w", filename, err)
		}

	}

	return nil
}

func findFiles(path string) ([]string, error) {
	var err error
	var files []string

	if path, err = filepath.Abs(path); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext == ".yaml" || ext == ".yml" || ext == ".cron" {
			files = append(files, filepath.Join(path, name))
		}
	}

	return files, err
}
