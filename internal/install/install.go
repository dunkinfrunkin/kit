package install

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dunkinfrunkin/kit/internal/detect"
)

type Item struct {
	Namespace string
	Type      string // "skill", "hook", "config"
	Name      string
	Content   []byte
}

type Options struct {
	Tools   []detect.Tool
	Project bool
	Target  string
}

func Install(item Item, opts Options) error {
	tools, err := resolveTools(opts)
	if err != nil {
		return err
	}

	for _, tool := range tools {
		var installErr error
		switch item.Type {
		case "skill":
			installErr = installSkill(item, tool, opts.Project)
		case "hook":
			installErr = installHook(item, tool, opts.Project)
		case "config":
			installErr = installConfig(item, tool, opts.Project)
		default:
			return fmt.Errorf("unknown item type: %q", item.Type)
		}
		if installErr != nil {
			return fmt.Errorf("install %s for %s: %w", item.Type, tool, installErr)
		}
	}
	return nil
}

func Uninstall(itemType, name string, opts Options) error {
	tools, err := resolveTools(opts)
	if err != nil {
		return err
	}

	for _, tool := range tools {
		var uninstallErr error
		switch itemType {
		case "skill":
			uninstallErr = uninstallSkill(name, tool, opts.Project)
		case "hook":
			uninstallErr = uninstallHook(name, tool, opts.Project)
		case "config":
			uninstallErr = uninstallConfig(name, tool, opts.Project)
		default:
			return fmt.Errorf("unknown item type: %q", itemType)
		}
		if uninstallErr != nil {
			return fmt.Errorf("uninstall %s for %s: %w", itemType, tool, uninstallErr)
		}
	}
	return nil
}

func resolveTools(opts Options) ([]detect.Tool, error) {
	if opts.Target != "" {
		t, err := detect.ParseTool(opts.Target)
		if err != nil {
			return nil, err
		}
		return []detect.Tool{t}, nil
	}
	if len(opts.Tools) > 0 {
		return opts.Tools, nil
	}
	return detect.DetectTools(), nil
}

func installSkill(item Item, tool detect.Tool, project bool) error {
	switch tool {
	case detect.Claude, detect.Codex:
		dir := filepath.Join(toolDir(tool, project), "skills", item.Name)
		return writeFile(filepath.Join(dir, "SKILL.md"), item.Content)
	case detect.Cursor:
		return nil
	}
	return nil
}

func installHook(item Item, tool detect.Tool, project bool) error {
	switch tool {
	case detect.Claude:
		return mergeJSONHook(item, tool, project)
	case detect.Codex:
		return appendCodexHook(item, tool, project)
	case detect.Cursor:
		return nil
	}
	return nil
}

func installConfig(item Item, tool detect.Tool, project bool) error {
	switch tool {
	case detect.Claude:
		path := filepath.Join(toolDir(tool, project), "CLAUDE.md")
		return appendMarked(path, item.Name, item.Content)
	case detect.Codex:
		path := filepath.Join(toolDir(tool, project), "AGENTS.md")
		return appendMarked(path, item.Name, item.Content)
	case detect.Cursor:
		dir := filepath.Join(toolDir(tool, project), "rules")
		content := fmt.Sprintf("---\ndescription: %s\nalwaysApply: true\n---\n%s\n", item.Name, string(item.Content))
		return writeFile(filepath.Join(dir, item.Name+".mdc"), []byte(content))
	}
	return nil
}

func uninstallSkill(name string, tool detect.Tool, project bool) error {
	switch tool {
	case detect.Claude, detect.Codex:
		dir := filepath.Join(toolDir(tool, project), "skills", name)
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			return err
		}
	case detect.Cursor:
		// no-op
	}
	return nil
}

func uninstallHook(name string, tool detect.Tool, project bool) error {
	switch tool {
	case detect.Claude:
		path := filepath.Join(toolDir(tool, project), "settings.json")
		return removeJSONHook(path, name)
	case detect.Codex:
		path := filepath.Join(toolDir(tool, project), "config.toml")
		return removeMarked(path, name)
	case detect.Cursor:
		// no-op
	}
	return nil
}

func uninstallConfig(name string, tool detect.Tool, project bool) error {
	switch tool {
	case detect.Claude:
		path := filepath.Join(toolDir(tool, project), "CLAUDE.md")
		return removeMarked(path, name)
	case detect.Codex:
		path := filepath.Join(toolDir(tool, project), "AGENTS.md")
		return removeMarked(path, name)
	case detect.Cursor:
		path := filepath.Join(toolDir(tool, project), "rules", name+".mdc")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func toolDir(tool detect.Tool, project bool) string {
	if project {
		return detect.ProjectDir(tool)
	}
	return detect.GlobalDir(tool)
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

func mergeJSONHook(item Item, tool detect.Tool, project bool) error {
	path := filepath.Join(toolDir(tool, project), "settings.json")

	existing := make(map[string]interface{})
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	var hookData map[string]interface{}
	if err := json.Unmarshal(item.Content, &hookData); err != nil {
		return fmt.Errorf("parse hook content: %w", err)
	}

	hooks, _ := existing["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}
	for k, v := range hookData {
		hooks[k] = v
	}
	existing["hooks"] = hooks

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, out)
}

func appendCodexHook(item Item, tool detect.Tool, project bool) error {
	path := filepath.Join(toolDir(tool, project), "config.toml")
	return appendMarked(path, item.Name, item.Content)
}

func removeJSONHook(path, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return nil
	}
	delete(hooks, name)
	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, out)
}

func markerStart(name string) string {
	return fmt.Sprintf("<!-- kit:%s -->", name)
}

func markerEnd(name string) string {
	return fmt.Sprintf("<!-- /kit:%s -->", name)
}

func appendMarked(filePath, name string, content []byte) error {
	existing, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	start := markerStart(name)
	end := markerEnd(name)
	section := start + "\n" + string(content) + "\n" + end + "\n"

	if bytes.Contains(existing, []byte(start)) {
		replaced := replaceSection(string(existing), start, end, section)
		return writeFile(filePath, []byte(replaced))
	}

	var buf bytes.Buffer
	if len(existing) > 0 {
		buf.Write(existing)
		if !bytes.HasSuffix(existing, []byte("\n")) {
			buf.WriteByte('\n')
		}
	}
	buf.WriteString(section)
	return writeFile(filePath, buf.Bytes())
}

func removeMarked(filePath, name string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	start := markerStart(name)
	end := markerEnd(name)

	if !bytes.Contains(data, []byte(start)) {
		return nil
	}

	replaced := replaceSection(string(data), start, end, "")
	replaced = strings.TrimRight(replaced, "\n") + "\n"
	if strings.TrimSpace(replaced) == "" {
		return os.Remove(filePath)
	}
	return writeFile(filePath, []byte(replaced))
}

func replaceSection(content, startMarker, endMarker, replacement string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}
	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx == -1 {
		return content
	}
	endIdx = startIdx + endIdx + len(endMarker)
	// Consume trailing newline if present.
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}
	return content[:startIdx] + replacement + content[endIdx:]
}
