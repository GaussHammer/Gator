package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func home() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	full_url := home + "/.gatorconfig.json"
	return full_url, nil
}

func Read() (Config, error) {
	var config Config
	full_url, _ := home()
	file_content, err := os.Open(full_url)
	if err != nil {
		return config, err
	}
	defer file_content.Close()

	result, err := io.ReadAll(file_content)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(result, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	full_url, _ := home()

	err = os.WriteFile(full_url, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
