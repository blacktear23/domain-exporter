package main

import (
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	fname              string
	CollectDuration    int      `yaml:"collect_duration"`
	CertificateDomains []string `yaml:"certificate_domains"`
	WhoisDomains       []string `yaml:"whois_domains"`
	lock               sync.RWMutex
}

func NewConfig(fname string) (*Config, error) {
	cfg := &Config{
		fname:              fname,
		CollectDuration:    3600,
		CertificateDomains: []string{},
		WhoisDomains:       []string{},
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

func (c *Config) GetDuration() time.Duration {
	return time.Second * time.Duration(c.CollectDuration)
}
