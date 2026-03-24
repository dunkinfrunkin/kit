package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dunkinfrunkin/kit/internal/config"
)

type InstalledItem struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Version   int    `json:"version"`
}

func trackingPath() string {
	return filepath.Join(config.Dir(), "installed.json")
}

func loadTracking() ([]InstalledItem, error) {
	data, err := os.ReadFile(trackingPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []InstalledItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func saveTracking(items []InstalledItem) error {
	if err := os.MkdirAll(filepath.Dir(trackingPath()), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackingPath(), data, 0644)
}

func trackInstall(namespace, itemType, name string, version int) error {
	items, err := loadTracking()
	if err != nil {
		return err
	}

	for i, item := range items {
		if item.Namespace == namespace && item.Type == itemType && item.Name == name {
			items[i].Version = version
			return saveTracking(items)
		}
	}

	items = append(items, InstalledItem{
		Namespace: namespace,
		Type:      itemType,
		Name:      name,
		Version:   version,
	})
	return saveTracking(items)
}

func trackUninstall(namespace, itemType, name string) error {
	items, err := loadTracking()
	if err != nil {
		return err
	}

	filtered := items[:0]
	for _, item := range items {
		if item.Namespace == namespace && item.Type == itemType && item.Name == name {
			continue
		}
		filtered = append(filtered, item)
	}
	return saveTracking(filtered)
}
