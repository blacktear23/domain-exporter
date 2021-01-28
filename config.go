package main

import (
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type RequestConfig struct {
	Host    string   `yaml:"host"`
	Domains []string `yaml:"domains"`
	Path    string   `yaml:"path"`
	Https   bool     `yaml:"https"`
}

type Config struct {
	fname              string
	CollectDuration    int             `yaml:"collect_duration"`
	CertificateDomains []string        `yaml:"certificate_domains"`
	WhoisDomains       []string        `yaml:"whois_domains"`
	ResolveDomains     []string        `yaml:"resolve_domains"`
	RequestDomains     []RequestConfig `yaml:"request_domains"`
	lock               sync.RWMutex
}

func NewConfig(fname string) (*Config, error) {
	cfg := &Config{
		fname:              fname,
		CollectDuration:    3600,
		CertificateDomains: []string{},
		WhoisDomains:       []string{},
		ResolveDomains:     []string{},
		RequestDomains:     []RequestConfig{},
	}
	err := cfg.Reload()
	return cfg, err
}

func (c *Config) Reload() error {
	file, err := os.Open(c.fname)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := yaml.NewDecoder(file)
	cfg := &Config{}
	err = dec.Decode(cfg)
	if err != nil {
		return err
	}
	c.lock.Lock()
	if cfg.CollectDuration > 60 {
		c.CollectDuration = cfg.CollectDuration
	}
	c.CertificateDomains = cfg.CertificateDomains
	c.WhoisDomains = cfg.WhoisDomains
	c.ResolveDomains = cfg.ResolveDomains
	c.RequestDomains = cfg.RequestDomains
	c.lock.Unlock()
	return nil
}

func (c *Config) GetCertificateDomains() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.CertificateDomains
}

func (c *Config) GetWhoisDomains() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.WhoisDomains
}

func (c *Config) GetResolveDomains() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.ResolveDomains
}

func (c *Config) GetRequestDomains() []RequestConfig {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.RequestDomains
}

func (c *Config) GetDuration() time.Duration {
	return time.Second * time.Duration(c.CollectDuration)
}
