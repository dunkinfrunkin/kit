package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Credentials struct {
	Server string `json:"server"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kit")
}

func credentialsPath() string {
	return filepath.Join(Dir(), "credentials")
}

func Load() (*Credentials, error) {
	data, err := os.ReadFile(credentialsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in: %s does not exist", credentialsPath())
		}
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials file: %w", err)
	}
	return &creds, nil
}

func Save(creds *Credentials) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsPath(), data, 0600)
}

func Remove() error {
	err := os.Remove(credentialsPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
