package core

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
)

var (
	DefaultEnvironmentVariables = map[string]string{
		"PATH":    "/bin:/usr/bin:/sbin:/usr/sbin:/bin/caca",
		"CONSOLE": "/dev/console",
	}
)

type ServiceManager struct {
	config *Config
}

type SortByPosition []*Service

func (a SortByPosition) Len() int           { return len(a) }
func (a SortByPosition) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByPosition) Less(i, j int) bool { return a[i].Position < a[j].Position }

func (s *ServiceManager) SortServicesByPosition() []*Service {
	services := s.config.GetServices()
	sort.Sort(SortByPosition(services))
	return services
}

func NewServiceManager(configPath string) (*ServiceManager, error) {
	config, err := NewConfigFromFile(configPath)

	if err != nil {
		return nil, err
	}

	return &ServiceManager{config: config}, nil
}

func (s *ServiceManager) HandleInterrupts() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	switch <-ch {
	case os.Interrupt:
		s.Stop()
	case syscall.SIGTERM:
		s.Stop()
	}

}

func (s *ServiceManager) SetEnvironment() error {
	for variable, value := range DefaultEnvironmentVariables {
		if err := os.Setenv(variable, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceManager) StartBanner() {
	fmt.Println(`☁ Tormenta Cloud OS -> Init ☁`)
}

func (s *ServiceManager) Start() error {
	s.StartBanner()

	err := s.SetEnvironment()
	if err != nil {
		return err
	}

	for _, service := range s.SortServicesByPosition() {
		go service.Start(s.config.Timeout)
	}

	s.HandleInterrupts()
	return nil
}

func (s *ServiceManager) Stop() {
	for _, service := range s.config.GetServices() {
		fmt.Println(fmt.Sprintf("Stoping Service: [%s] -> Description: [%s]", service.Name,
			service.Description))
	}
}

func (s *ServiceManager) Restart() {
	for _, service := range s.config.GetServices() {
		fmt.Println(fmt.Sprintf("Stoping Service: [%s] -> Description: [%s]", service.Name,
			service.Description))
	}
}
