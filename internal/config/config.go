package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FilePrefix    = ".workos"
	FileExtension = "json"
	FileName      = FilePrefix + "." + FileExtension
)

type Config struct {
	ActiveApiKey string            `mapstructure:"active_api_key" json:"active_api_key"`
	ApiKeys      map[string]ApiKey `mapstructure:"api_keys"       json:"api_keys"`
}

type ApiKey struct {
	Name        string `mapstructure:"name"        json:"name"`
	Value       string `mapstructure:"value"       json:"value"`
	Environment string `mapstructure:"environment" json:"environment"`
	Endpoint    string `mapstructure:"endpoint"    json:"endpoint"`
}

func (c Config) Write() error {
	fileContents, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	err = os.WriteFile(homeDir+"/"+FileName, fileContents, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadConfig() *Config {
	// Look for .warrant.json in HOME dir and create an empty version if it doesn't exist
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	_, err = os.Stat(homeDir + "/" + FileName)
	if errors.Is(err, fs.ErrNotExist) {
		emptyJson := []byte("{}")
		err = os.WriteFile(homeDir+"/"+FileName, emptyJson, 0644)
		cobra.CheckErr(err)
	}

	// Load config from ~/.workos.json
	viper.AddConfigPath(homeDir)
	viper.SetConfigType(FileExtension)
	viper.SetConfigName(FilePrefix)
	viper.AutomaticEnv() // read in environment variables that match
	err = viper.ReadInConfig()
	cobra.CheckErr(err)

	// Unmarshal config & set warrant client vals
	var config Config
	err = viper.Unmarshal(&config)
	cobra.CheckErr(err)
	return &config
}
