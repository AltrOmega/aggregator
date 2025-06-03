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

func (cfg Config) SetUser(username string) (Config, error) {
	cfg.Current_user_name = username
	err := Write(cfg)
	return cfg, err
}

func (cfg Config) SetDbUrl(dbUrl string) (Config, error) {
	cfg.Db_url = dbUrl
	err := Write(cfg)
	return cfg, err
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
		fmt.Println("Error marshalling JSON: ", err)
		return nil, err
	}
	return jsonData, nil
}

func Read() (Config, error) {
	file_path, err := GetConfigFilePath()
	if err != nil {
		fmt.Println("GetConfigFilePath error: ", err)
		return Config{}, err
	}

	file, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Println("File read error: ", err)
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal([]byte(file), &config)
	if err != nil {
		fmt.Println("Json unmarshal error: ", err)
		return Config{}, err
	}

	return config, nil
}

func Write(cfg Config) error {
	file_path, err := GetConfigFilePath()
	if err != nil {
		fmt.Println("GetConfigFilePath error: ", err)
		return err
	}

	jsonData, err := cfg.AsByte()
	if err != nil {
		fmt.Println("Error Byteing cfg: ", err)
		return err
	}

	file, err := os.Create(file_path)
	if err != nil {
		fmt.Println("Error opening file to write: ", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file: ", err)
		return err
	}

	return nil
}
