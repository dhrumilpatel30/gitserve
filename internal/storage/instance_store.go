package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Instance represents a running or detached gitserve instance.
type Instance struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	Path       string    `json:"path"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"startTime"`
	StopTime   time.Time `json:"stopTime,omitempty"` // Time the instance was stopped or entered a terminal state
	LogPath    string    `json:"logPath"`
	GitServeID string    `json:"gitserveId"`
}

// InstanceStore defines the interface for managing gitserve instances.
type InstanceStore interface {
	AddInstance(instance Instance) error
	GetInstanceByID(id string) (Instance, bool, error)
	GetAllInstances() ([]Instance, error)
	UpdateInstance(id string, updatedInstance Instance) error
	DeleteInstance(id string) error
}

const instancesFile = "gitserve_instances.json"

// jsonInstanceStore is a file-based implementation of InstanceStore using JSON.
type jsonInstanceStore struct {
	storagePath    string
	instances      map[string]Instance
	instancesMutex sync.RWMutex
}

// NewJSONInstanceStore creates and initializes a new JSON-based InstanceStore.
func NewJSONInstanceStore(dataDirPath string) (InstanceStore, error) {
	store := &jsonInstanceStore{
		storagePath: dataDirPath,
		instances:   make(map[string]Instance),
	}

	if err := ensureDirExists(store.storagePath); err != nil {
		return nil, fmt.Errorf("failed to ensure storage directory exists: %w", err)
	}

	if err := store.loadInstances(); err != nil {
		// If loading fails (e.g. corrupted file), return the error.
		// The caller (e.g., CLI command) can then decide how to handle it (e.g., exit, or offer to reset).
		return nil, fmt.Errorf("failed to load instances: %w", err)
	}
	return store, nil
}

func ensureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0750) // rwxr-x---
	}
	return nil
}

// loadInstances reads the instances from the JSON file into memory.
func (s *jsonInstanceStore) loadInstances() error {
	s.instancesMutex.Lock()
	defer s.instancesMutex.Unlock()

	instancesFilePath := filepath.Join(s.storagePath, instancesFile)
	data, err := os.ReadFile(instancesFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, initialize with an empty map (this is not an error)
			s.instances = make(map[string]Instance)
			return nil
		}
		return fmt.Errorf("error reading instances file %s: %w", instancesFilePath, err)
	}

	if len(data) == 0 {
		// File is empty, initialize with an empty map
		s.instances = make(map[string]Instance)
		return nil
	}

	// Attempt to unmarshal. If it fails, reset to an empty map and return the error.
	if err = json.Unmarshal(data, &s.instances); err != nil {
		s.instances = make(map[string]Instance) // Reset to a clean state on unmarshal error
		return fmt.Errorf("error unmarshalling instances data from %s: %w. Store reset to empty.", instancesFilePath, err)
	}
	return nil
}

// saveInstances writes the current in-memory instances to the JSON file.
// It assumes that the caller (AddInstance, UpdateInstance, DeleteInstance)
// already holds the necessary (write) lock on s.instancesMutex.
func (s *jsonInstanceStore) saveInstances() error {
	instancesFilePath := filepath.Join(s.storagePath, instancesFile)
	fmt.Fprintf(os.Stderr, "[DEBUG] Attempting to save instances to: %s\n", instancesFilePath) // DEBUG PRINT

	data, err := json.MarshalIndent(s.instances, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] Error marshalling instances: %v\n", err) // DEBUG PRINT
		return fmt.Errorf("error marshalling instances: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Marshalled data length: %d bytes\n", len(data)) // DEBUG PRINT

	// Direct write for diagnostics
	if err = os.WriteFile(instancesFilePath, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] Error writing instances file %s: %v\n", instancesFilePath, err) // DEBUG PRINT
		return fmt.Errorf("error writing instances file %s: %w", instancesFilePath, err)
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Successfully wrote instances file: %s\n", instancesFilePath) // DEBUG PRINT
	return nil
}

// AddInstance adds a new instance to the store.
func (s *jsonInstanceStore) AddInstance(instance Instance) error {
	s.instancesMutex.Lock()
	defer s.instancesMutex.Unlock()

	if _, exists := s.instances[instance.ID]; exists {
		return fmt.Errorf("instance with ID '%s' already exists", instance.ID)
	}
	s.instances[instance.ID] = instance
	return s.saveInstances()
}

// GetInstanceByID retrieves a specific instance by its ID.
func (s *jsonInstanceStore) GetInstanceByID(id string) (Instance, bool, error) {
	s.instancesMutex.RLock()
	defer s.instancesMutex.RUnlock()

	instance, found := s.instances[id]
	return instance, found, nil
}

// GetAllInstances returns a slice of all stored instances.
func (s *jsonInstanceStore) GetAllInstances() ([]Instance, error) {
	s.instancesMutex.RLock()
	defer s.instancesMutex.RUnlock()

	all := make([]Instance, 0, len(s.instances))
	for _, instance := range s.instances {
		all = append(all, instance)
	}
	return all, nil
}

// UpdateInstance modifies an existing instance in the store.
func (s *jsonInstanceStore) UpdateInstance(id string, updatedInstance Instance) error {
	s.instancesMutex.Lock()
	defer s.instancesMutex.Unlock()

	if _, exists := s.instances[id]; !exists {
		return fmt.Errorf("instance with ID '%s' not found for update", id)
	}
	s.instances[id] = updatedInstance
	return s.saveInstances()
}

// DeleteInstance removes an instance from the store by its ID.
func (s *jsonInstanceStore) DeleteInstance(id string) error {
	s.instancesMutex.Lock()
	defer s.instancesMutex.Unlock()

	if _, exists := s.instances[id]; !exists {
		return fmt.Errorf("instance with ID '%s' not found for delete", id)
	}
	delete(s.instances, id)
	return s.saveInstances()
}
