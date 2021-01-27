package main

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	config *Config
}

func NewCollector(cfg *Config) *Collector {
	return &Collector{
		config: cfg,
	}
}

func decodeStatus(status string) float64 {
	switch status {
	case "Error":
		return 0
	case "OK":
		return 1
	default:
		return 0
	}
}

func (c *Collector) collectCertificates() {
	log.Println("Collect Certificates")
	checker := NewCertificatesChecker(c.config.GetCertificateDomains())
	results := checker.Check()
	for domain, result := range results {
		DomainCertificateStatus.With(prometheus.Labels{"domain": domain}).Set(decodeStatus(result.Status))
		DomainCertificateExpireDays.With(prometheus.Labels{"domain": domain}).Set(float64(result.ExpireDays))
	}
	log.Println("Collect Certificates Finish")
}

func (c *Collector) collectDomains() {
	log.Println("Collect Whois Informations")
	checker := NewWhoisChecker(c.config.GetWhoisDomains())
	results := checker.Check()
	for domain, result := range results {
		DomainWhoisStatus.With(prometheus.Labels{"domain": domain}).Set(decodeStatus(result.Status))
		DomainWhoisExpireDays.With(prometheus.Labels{"domain": domain}).Set(float64(result.ExpireDays))
	}
	log.Println("Collect Whois Informations Finish")
}

func (c *Collector) collectResolves() {
	log.Println("Collect Resolve Informations")
	checker := NewResolveChecker(c.config.GetResolveDomains())
	results := checker.Check()
	for domain, result := range results {
		DomainResolveStatus.With(prometheus.Labels{"domain": domain}).Set(decodeStatus(result.Status))
		DomainResolveIPs.With(prometheus.Labels{"domain": domain}).Set(float64(len(result.IPs)))
	}
}

func (c *Collector) CollectOnce() {
	go c.collectCertificates()
	go c.collectDomains()
	go c.collectResolves()
}

func (c *Collector) Start() {
	for {
		c.CollectOnce()
		sleepSecs := c.config.GetDuration()
		time.Sleep(sleepSecs)
	}
}
