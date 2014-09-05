package core

import (
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

type Service struct {
	ConfigPath   string
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Position     int      `yaml:"position"`
	Timeout      int      `yaml:"timeout"`
	StartSteps   []string `yaml:"start"`
	StopSteps    []string `yaml:"stop"`
	RestartSteps []string `yaml:"restart"`
	Block        bool     `yaml:"block"`
	CheckFS      []string `yaml:"checkfs"`
}

func ValidateService(s *Service) error {

	if s.Description == "" {
		s.Description = "default service description"
	}

	if s.Name == "" {
		return errors.New("name not defined")
	}

	if len(s.StartSteps) == 0 {
		return errors.New("start steps not defined")
	}

	if len(s.StopSteps) == 0 {
		return errors.New("stop steps not defined")
	}

	return nil
}

func NewService(path string) (*Service, error) {
	service := Service{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("Cannot open configuration file")
	}

	readed, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("Cannot read configuration file")
	}

	err = goyaml.Unmarshal([]byte(readed), &service)
	if err != nil {
		return nil, fmt.Errorf("Error parsing service configuration: %v", err)
	}

	service.ConfigPath = path

	err = ValidateService(&service)
	if err != nil {
		return nil, fmt.Errorf("Invalid service definition (path:%s): %v", path, err)
	}

	return &service, nil
}

func (s *Service) IsBlocking() bool {
	return s.Block == true
}

func (s *Service) String() string {
	return fmt.Sprintf("Service: %s (definition:%s)", s.Name, s.ConfigPath)
}

func (s *Service) Start(timeout int) {

	for _, command := range s.StartSteps {

		cmd := exec.Command("/bin/bash", "-c", command)
		if !s.IsBlocking() {
			if err := cmd.Start(); err != nil {
				log.Fatal("Cannot start service: %s , error: %v", s.Name, err)
			}
			log.Printf("%s, [started]", s)
			return
		}

		done := make(chan error)
		go func() {
			done <- cmd.Run()
		}()

		if s.Timeout > timeout {
			timeout = s.Timeout
		}

		select {
		case <-time.After(time.Duration(timeout) * time.Second):
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			<-done
			log.Printf("%s, [failed], reason: timeout after %d secs", s, timeout)
		case err := <-done:
			if err != nil {
				log.Printf("%s, [failed], reason: %v", s, err)
			} else {
				log.Printf("%s, [started]", s)
			}

		}

		defer close(done)
	}
}
