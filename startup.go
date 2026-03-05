package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (e *editor) loadFromArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) > 1 {
		return errors.New("only one file can be opened at a time")
	}
	return e.openFile(args[0])
}

func (e *editor) openFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("directories are not supported")
	}

	if !isSupportedTextFilePath(path) {
		return fmt.Errorf("unsupported file type %q (allowed: .txt, .md)", strings.ToLower(filepath.Ext(path)))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	text := normalizeClipboardText(string(data))
	parts := strings.Split(text, "\n")
	if len(parts) == 0 {
		parts = []string{""}
	}

	lines := make([][]rune, 0, len(parts))
	for _, p := range parts {
		lines = append(lines, []rune(p))
	}
	if len(lines) == 0 {
		lines = [][]rune{{}}
	}

	e.lines = lines
	e.row = 0
	e.col = 0
	e.scroll = 0
	e.normalPending = 0
	e.flashLine = -1
	e.filePath = path
	e.commandLineActive = false
	e.commandLine = e.commandLine[:0]
	e.commandCol = 0
	e.commandError = ""
	e.setMode(modeNormal)
	return nil
}

func isSupportedTextFilePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".txt" || ext == ".md"
}
