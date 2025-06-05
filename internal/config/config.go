package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func (cfg *Config) SetUser(username string) error {
	cfg.Current_user_name = username
	err := Write(*cfg)
	return err
}

func (cfg *Config) SetDbUrl(dbUrl string) error {
	cfg.Db_url = dbUrl
	err := Write(*cfg)
	return err
}

const ConfigFileName = `.gatorconfig.json`

func GetConfigFilePath() (string, error) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return home_dir + "/" + ConfigFileName, nil
}

func (cfg Config) AsByte() ([]byte, error) {
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}
	return jsonData, nil
}

func Read() (Config, error) {
	file_path, err := GetConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("error in GetConfigFilePath: %w", err)
	}

	file, err := os.ReadFile(file_path)
	if err != nil {
		return Config{}, fmt.Errorf("file read error: %w", err)
	}

	var config Config
	err = json.Unmarshal([]byte(file), &config)
	if err != nil {
		return Config{}, fmt.Errorf("json unmarshal error: %w", err)
	}

	return config, nil
}

func Write(cfg Config) error {
	file_path, err := GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("error in GetConfigFilePath: %w", err)
	}

	jsonData, err := cfg.AsByte()
	if err != nil {
		return fmt.Errorf("error Byteing cfg: %w", err)
	}

	file, err := os.Create(file_path)
	if err != nil {
		return fmt.Errorf("error opening file to write: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
