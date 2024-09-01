package utils

import (
	"os"
	"os/user"
	"path"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Infra struct {
		Path string `yaml:"path"`
	}
}

func RetrieveOrGenerateConfig(repo string) (Config, string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config_folder := path.Join(dir, BASE_CONFIG)
	config_file := path.Join(config_folder, "config.yml")

	if !FolderExists(config_folder) {
		err := os.MkdirAll(config_folder, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	if !FileExists(config_file) {
		err := os.WriteFile(config_file, []byte(""), 0755)
		if err != nil {
			panic(err)
		}
	}
	file, err := os.ReadFile(config_file)
	if err != nil {
		panic(err)
	}
	config := Config{}

	err = yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		panic(err)
	}
	return config, config_file
}

func SaveConfig(config *Config, config_file string) {
	file, err := yaml.Marshal(&config)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(config_file, file, 0755)
	if err != nil {
		panic(err)
	}
}
