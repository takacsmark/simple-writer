package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func (e *editor) saveBuffer(pathArg string) error {
	path := strings.TrimSpace(pathArg)
	if path == "" {
		path = e.filePath
	}
	if strings.TrimSpace(path) == "" {
		return errors.New("no file name")
	}
	if !isSupportedTextFilePath(path) {
		return fmt.Errorf("unsupported file type (allowed: .txt, .md)")
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return errors.New("target path is a directory")
	}

	content := e.bufferText()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	e.filePath = path
	return nil
}

func (e *editor) bufferText() string {
	if len(e.lines) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e.lines))
	for _, line := range e.lines {
		parts = append(parts, string(line))
	}
	return strings.Join(parts, "\n")
}
