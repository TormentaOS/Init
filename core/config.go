package core

import (
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	DefaultStartMessage = "Starting initialize"
	DefaultStopMessage  = "Stopping initialize"
)

type Config struct {
	Path          string
	ServicesArr   []*Service
	Timeout       int    `yaml:"timeout"`
	StartMessage  string `yaml:"start"`
	StopMessage   string `yaml:"stop"`
	ServicesField string `yaml:"services"`
}

//NewConfigFromFile function loads a YAML file and returns
//a pointer to a newly create Configuration struct
func NewConfigFromFile(path string) (*Config, error) {
	config := Config{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Cannot find configuration:%s", path))
	}

	readed, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("Cannot read configuration file")
	}

	err = goyaml.Unmarshal(readed, &config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read configuration: %v", err))
	} else {
		config.Path = path
	}

	err = ValidateConfig(&config)
	if err != nil {
		return nil, err
	}

	path, err = config.GetServicesPath()
	if err != nil {
		return nil, err
	}

	servicesPaths, err := filepath.Glob(path)
	if err != nil {
		return nil, errors.New("cannot find defined services")
	}

	for _, servicePath := range servicesPaths {
		service, err := NewService(servicePath)
		if err != nil {
			return nil, err
		}
		config.ServicesArr = append(config.ServicesArr, service)
	}

	return &config, nil
}

func (c *Config) GetPath() (string, error) {
	return c.Path, nil
}

func (c *Config) GetServices() []*Service {
	return c.ServicesArr
}

func (c *Config) GetStartMessage() (string, error) {
	if c.StartMessage == "" {
		c.StartMessage = DefaultStartMessage
	}
	return c.StartMessage, nil
}

func (c *Config) GetStopMessage() (string, error) {
	if c.StartMessage == "" {
		c.StopMessage = DefaultStartMessage
	}
	return c.StopMessage, nil
}

func (c *Config) GetServicesPath() (string, error) {
	if len(c.ServicesField) == 0 {
		return "", errors.New("not defined services found")
	}
	return c.ServicesField, nil
}

func ValidateConfig(c *Config) error {
	var err error

	if _, err = c.GetPath(); err != nil {
		return err
	}

	if _, err = c.GetStartMessage(); err != nil {
		return err
	}

	if _, err = c.GetStopMessage(); err != nil {
		return err
	}

	if _, err = c.GetServicesPath(); err != nil {
		return err
	}
	return nil
}
