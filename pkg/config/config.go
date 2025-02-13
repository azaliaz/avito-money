package config

import (
	"os"

	"github.com/caarlos0/env/v10"

	"gopkg.in/yaml.v3"
)

// ReadConfig reads configuration from file or from env variables if configFile set to "none"
// use envPrefix,yaml tags like `envPrefix:"DB_" yaml:"storage"`.
func ReadConfig(configFile string, cfg any) error {
	if configFile != "" && configFile != "none" {
		err := parseYaml(configFile, cfg)
		if err != nil {
			return err
		}
	} else {
		err := parseEnv(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseEnv(cfg any) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}

func parseYaml(file string, cfg any) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	yamlDecoder := yaml.NewDecoder(f)
	err = yamlDecoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func CreateYaml(outputFileName string, cfg any) error {
	f, err := os.Create(outputFileName)
	if err != nil {
		return err
	}

	defer f.Close()

	yamlDecoder := yaml.NewEncoder(f)
	err = yamlDecoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
