package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Schedules map[string]*job `json:"schedules" yaml:"schedules"`
}

func loadSetting(v any, filename string) error {
	if content, err := os.ReadFile(filename); err != nil {
		return fmt.Errorf("read settings file error: %w", err)
	} else {
		if err = yaml.Unmarshal(content, v); err != nil {
			return fmt.Errorf("unmarshal settings file \"%s\" error: %w", filename, err)
		}
	}
	return nil
}

func findYamls(path string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, filepath.Join(path, name))
		}
	}

	return files, err
}
