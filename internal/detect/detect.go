package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool int

const (
	Claude Tool = iota
	Codex
	Cursor
)

var allTools = []Tool{Claude, Codex, Cursor}

var toolNames = map[Tool]string{
	Claude: "claude",
	Codex:  "codex",
	Cursor: "cursor",
}

func (t Tool) String() string {
	return toolNames[t]
}

func DetectTools() []Tool {
	home, _ := os.UserHomeDir()
	var found []Tool
	for _, t := range allTools {
		dir := filepath.Join(home, "."+t.String())
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			found = append(found, t)
		}
	}
	if len(found) == 0 {
		return []Tool{Claude, Codex, Cursor}
	}
	return found
}

func GlobalDir(t Tool) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+t.String())
}

func ProjectDir(t Tool) string {
	return "." + t.String()
}

func ParseTool(s string) (Tool, error) {
	lower := strings.ToLower(s)
	for tool, name := range toolNames {
		if name == lower {
			return tool, nil
		}
	}
	return 0, fmt.Errorf("unknown tool: %q (valid: claude, codex, cursor)", s)
}
