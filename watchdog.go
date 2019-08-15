package watchdog

import (
	"fmt"
	"sync"
	"time"
)

// WatchdogService is an interface to a service that forms part of the watchdog system
type WatchdogService interface {
	Name() string
	Whack()
	Check() bool
}

// WatchdogSystem allows one to watch a set of services
type WatchdogSystem interface {
	Add(s Service)
	Remove(s Service)
	Check() error
	Watch(period time.Duration) error
	Terminate()
}

// timeWatchdogService implements a time-based watchdog check
type timeWatchdogService struct {
	name     string
	duration time.Duration
	expireAt time.Time
}

// TimeWatchdogService returns a time-based watchdog service
func TimeWatchdogService(name string, duration time.Duration) WatchdogService {
	s := timeWatchdogService{
		name:     name,
		duration: duration,
	}

	s.Whack()

	return &s
}

// Name returns the name of the service
func (s *timeWatchdogService) Name() string {
	return s.name
}

// Whatck resets the service for its configured duration
func (s *timeWatchdogService) Whack() {
	s.expireAt = time.Now().Add(s.duration)
}

// Check checks if the service Whack time has expired
func (s *timeWatchdogService) Check() bool {
	return time.Now().After(s.expireAt)
}

// mapWatchdogSystem implements a system of managing watchdog services using a map
type mapWatchdogSystem struct {
	terminated bool
	mutex      *sync.Mutex
	services   map[string]WatchdogService
}

// MapWatchdogSystem returns a new watchdog system implemented using maps
func MapWatchdogSystem() Watchdog {
	w := timeWatchdog{
		terminated: false,
		mutex:      new(sync.Mutex),
		services:   make(map[string]WatchdogService),
	}

	return &w
}

// Add adds a service to the list of services checked
func (w *mapWatchdogSystem) Add(s Service) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.timeouts[s.Name()] = s
}

// Remove removes a service from the list of services checked
func (w *mapWatchdogSystem) Remove(s Service) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	delete(w.timeouts, s.Name())
}

// Check checks all services for faults
func (w *mapWatchdogSystem) Check() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	errors := ""

	for name, service := range w.services {
		if !service.Check() {
			if errors != "" {
				errors = errors + ", "
			}

			errors += name
		}
	}

	if errors != "" {
		return fmt.Errorf("Watchdog timed out on the following services: %s", errors)
	}

	return nil
}

// Watch continually watches the services until it terminates or there is a failure. This should run in a goroutine
func (w *mapWatchdogSystem) Watch(period time.Duration) error {
	for {
		if w.terminated {
			return nil
		}

		err := w.Check()
		if err != nil {
			return err
		}

		time.Sleep(period)
	}
}

// Terminate terminates the Watch method.
func (w *mapWatchdogSystem) Terminate() {
	w.terminated = true
}
