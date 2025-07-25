package nats

import (
	"errors"
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type NatConfig struct {
	Selector string  `yaml:"selector"`
	Name     *string `yaml:"name,omitempty"`  // Optional field
	Queue    *string `yaml:"queue,omitempty"` // Optional field
	Subject  string  `yaml:"subject"`
	URL      string  `yaml:"url"`
}

func (r *NatConfig) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("selector: %s, name: %v, queue: %v, subject: %s, url: %s", r.Selector, r.Name, r.Queue, r.Subject, r.URL)
}

// if the queue is set then
func (r *NatConfig) JetStreamEnabled() bool {
	if r.Queue != nil && r.Name != nil {
		return true
	}

	log.Println(fmt.Sprintf("JetStream is disabled, js name: %v, queue: %v, subject: %v", r.Name, r.Queue, r.Subject))
	return false
}

type MessageConfig struct {
	Routes []NatConfig `yaml:"routes"`
}

type Config struct {
	MessageConfig MessageConfig `yaml:"messageConfig"`

	selectorMap map[string]*NatConfig
	subjectMap  map[string]*NatConfig
}

// LoadConfig reads the YAML file and builds hash maps for fast route lookups
func LoadConfig(configPath string) (*Config, error) {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	// Build the hash maps for quick lookups
	config.BuildRouteMaps()

	return &config, nil
}

func LoadConfigFromEnv() (*Config, error) {
	// load the nats routing table from environment variable
	config, err := LoadConfig(utils.StringFromEnvWithDefault("ROUTING_FILE", "../routing-nats.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load routing config: %v", err)
	}
	return config, nil
}

// BuildRouteMaps builds hash maps for selector and subject for fast lookups
func (c *Config) BuildRouteMaps() {
	c.selectorMap = make(map[string]*NatConfig)
	c.subjectMap = make(map[string]*NatConfig)

	for i := range c.MessageConfig.Routes {
		route := &c.MessageConfig.Routes[i]
		c.selectorMap[route.Selector] = route
		c.subjectMap[route.Subject] = route
	}
}

// FindRouteBySelector finds a route by its selector using the hash map
func (c *Config) FindRouteBySelector(selector string) (*NatConfig, error) {
	route, found := c.selectorMap[selector]
	if !found {
		return nil, fmt.Errorf("route not found by selector %v", selector)
	}
	return route, nil
}

// FindRouteBySubject finds a route by its subject using the hash map
func (c *Config) FindRouteBySubject(subject string) (*NatConfig, error) {
	route, found := c.subjectMap[subject]
	if !found {
		return nil, errors.New("route not found by subject")
	}
	return route, nil
}
