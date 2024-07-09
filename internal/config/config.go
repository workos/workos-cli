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
	ActiveEnvironment string                 `mapstructure:"active_environment" json:"active_environment"`
	Environments      map[string]Environment `mapstructure:"environments"       json:"environments"`
}

type Environment struct {
	Endpoint string `mapstructure:"endpoint" json:"endpoint"`
	Name     string `mapstructure:"name"     json:"name"`
	Type     string `mapstructure:"type"     json:"type"`
	ApiKey   string `mapstructure:"api_key"    json:"api_key"`
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
	// Look for .workos.json in HOME dir and create an empty version if it doesn't exist
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
