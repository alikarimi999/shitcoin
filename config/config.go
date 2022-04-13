package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config interface {
	SaveConfig() error
	SetNodeID(id string)
	GetNodeID() string
}

type config struct {
	configName string
	configPath string
	configType string

	viper *viper.Viper
}

func NewConfig(path string) *config {

	s := strings.Split(path, string(os.PathSeparator))

	c := &config{}
	file := s[len(s)-1]
	fsplit := strings.Split(file, ".")
	if len(fsplit) == 1 {
		c.configType = "yaml"
	}
	c.configType = fsplit[len(fsplit)-1]
	c.configName = file
	c.configPath = strings.Join(s[:len(s)-1], string(os.PathSeparator))

	t := strings.Split(c.configName, ".")
	c.configType = t[len(t)-1]

	c.viper = viper.New()
	c.Creat()

	return c

}

func (c *config) Creat() {
	c.viper.SetConfigName(c.configName)
	c.viper.AddConfigPath(c.configPath)
	c.viper.SetConfigType(c.configType)

	if err := c.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.MkdirAll(c.configPath, os.ModePerm)
			os.Create(filepath.Join(c.configPath, c.configName))
		}
	}

}

func (c *config) SetNodeID(id string) {

	c.viper.SetDefault("NODE.ID", id)
}

func (c *config) GetNodeID() string {
	return c.viper.GetString("NODE.ID")
}

func (c *config) SaveConfig() error {
	return c.viper.WriteConfig()
}
