package core

import (
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	DefaultEnvironmentVariables = map[string]string{
		"PATH":    "/bin:/usr/bin:/sbin:/usr/sbin:/bin/caca",
		"CONSOLE": "/dev/console",
	}
)

type Service struct {
	ConfigPath       string
	NameField        string   `yaml:"name"`
	DescriptionField string   `yaml:"description"`
	PositionField    int      `yaml:"position"`
	StartField       []string `yaml:"start"`
	StopField        []string `yaml:"stop"`
	RestartField     []string `yaml:"restart"`
	BlockField       bool     `yaml:"block"`
	CheckFSField     []string `yaml:"checkfs"`
}

type ServiceManager struct {
	config *Config
}

type SortByPosition []*Service

func (a SortByPosition) Len() int           { return len(a) }
func (a SortByPosition) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByPosition) Less(i, j int) bool { return a[i].GetPosition() < a[j].GetPosition() }

func NewServiceManager(configPath string) (*ServiceManager, error) {
	config, err := NewConfigFromFile(configPath)

	if err != nil {
		return nil, err
	}

	return &ServiceManager{config: config}, nil
}

func (s *ServiceManager) Handle() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	switch <-ch {
	case os.Interrupt:
		s.Stop()
	case syscall.SIGTERM:
		s.Stop()
	}

}

func (s *ServiceManager) GetServicesByPosition() []*Service {
	services := s.config.GetServices()
	sort.Sort(SortByPosition(services))
	return services
}

func (s *ServiceManager) SetEnvironment() error {
	for variable, value := range DefaultEnvironmentVariables {
		if err := os.Setenv(variable, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceManager) PrintBanner() {
	fmt.Println(`☁ Tormenta Init - Cloud OS ☁`)
}

func (s *ServiceManager) Start() error {
	s.PrintBanner()

	err := s.SetEnvironment()
	if err != nil {
		return err
	}

	for _, service := range s.GetServicesByPosition() {
		go service.Start()
	}

	s.Handle()
	return nil
}

func (s *ServiceManager) Stop() {
	for _, service := range s.config.GetServices() {
		fmt.Println(fmt.Sprintf("Stoping Service: [%s] -> Description: [%s]", service.NameField,
			service.DescriptionField))
	}
}

func NewServiceFromFile(path string) (*Service, error) {
	service := Service{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("cannot open configuration file")
	}

	readed, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("cannot read configuration file")
	}

	err = goyaml.Unmarshal([]byte(readed), &service)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error parsing service configuration: %v", err))
	}

	service.ConfigPath = path

	err = ValidateService(&service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (s *Service) GetConfigPath() string {
	return s.ConfigPath
}

func (s *Service) GetPosition() int {
	return s.PositionField
}

func (s *Service) GetName() (string, error) {
	if s.NameField == "" {
		return "", errors.New("not defined name for service")
	}
	return s.NameField, nil
}

func (s *Service) GetCheckFS() ([]string, error) {
	return s.CheckFSField, nil
}

func (s *Service) GetDescription() (string, error) {
	if s.DescriptionField == "" {
		s.DescriptionField = "default service description"
	}
	return s.DescriptionField, nil
}

func (s *Service) GetStart() ([]string, error) {
	fmt.Println(s.StartField)
	if len(s.StartField) == 0 {
		return nil, errors.New("Not defined start steps for service " + s.NameField)
	}
	return s.StartField, nil
}

func (s *Service) IsBlocking() bool {
	return s.BlockField
}

func (s *Service) Start() {
	for _, command := range s.StartField {
		cmd := exec.Command(["/bin/bash", "-c", command])
		_, _ = cmd.StdoutPipe()
		if s.IsBlocking() {
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			select {
			case <-time.After(3 * time.Second):
				if err := cmd.Process.Kill(); err != nil {
					log.Fatal("failed to kill: ", err)
				}
				<-done // allow goroutine to exit
				log.Println("process killed for timeout")
			case err := <-done:
				if err != nil {
					log.Printf("process done with error = %v", err)
				}
				log.Printf("Started Service [%s]", s.NameField)

			}
		} else {
			cmd.Run()
			log.Printf("Started Service [%s]", s.NameField)

		}
	}
}

func ValidateService(s *Service) error {
	var err error

	if _, err = s.GetName(); err != nil {
		return err
	}

	if _, err = s.GetDescription(); err != nil {
		return err
	}

	if _, err = s.GetStart(); err != nil {
		return err
	}

	return nil
}
