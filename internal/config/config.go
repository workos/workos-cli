package config

import (
	"encoding/json"
	"errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"strings"
)

const (
	EnvVarPrefix       = "WORKOS"
	EnvVarHeadlessMode = "headless"
	FilePrefix         = ".workos"
	FileExtension      = "json"
	FileName           = FilePrefix + "." + FileExtension
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

// Creates an empty config file if it doesn't exist
func createEmptyConfigFile(dir string) {
	_, err := os.Stat(dir + "/" + FileName)
	if errors.Is(err, fs.ErrNotExist) {
		emptyJson := []byte("{}")
		err = os.WriteFile(dir+"/"+FileName, emptyJson, 0644)
		cobra.CheckErr(err)
	}
}

// Loads config values from environment variables if active environment is set to headless mode
// Supports overriding nested json keys with environment variables
// e.g. environments.headless.endpoint -> WORKOS_ENVIRONMENTS_HEADLESS_ENDPOINT
func loadEnvVarOverrides() {
	viper.SetEnvPrefix(EnvVarPrefix)
	// replace '.' in env var names with '_' to support overriding nested json keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// read in environment variables that match
	viper.AutomaticEnv()

	_ = viper.BindEnv("active_environment")
	activeEnvironment := viper.Get("active_environment")

	// Binds environment variables to nested json keys which allows unmarshalling into struct
	if activeEnvironment == EnvVarHeadlessMode {
		_ = viper.BindEnv("environments.headless.endpoint")
		_ = viper.BindEnv("environments.headless.type")
		_ = viper.BindEnv("environments.headless.name")
		_ = viper.BindEnv("environments.headless.api_key")
	}
}

func LoadConfig() *Config {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	createEmptyConfigFile(homeDir)

	// Load config from ~/.workos.json
	viper.AddConfigPath(homeDir)
	viper.SetConfigType(FileExtension)
	viper.SetConfigName(FilePrefix)

	loadEnvVarOverrides()

	err = viper.ReadInConfig()
	cobra.CheckErr(err)

	// Unmarshal config & set warrant client vals
	var config Config
	err = viper.Unmarshal(&config)
	cobra.CheckErr(err)
	return &config
}
