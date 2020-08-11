package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registry = prometheus.NewRegistry()

	DomainCertificateStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_certificate_status",
			Help: "Domain certificate status, 0 means error, 1 means OK.",
		},
		[]string{"domain"},
	)

	DomainCertificateExpireDays = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_certificate_expire_days",
			Help: "Domain certificate expire days.",
		},
		[]string{"domain"},
	)

	DomainWhoisStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_whois_status",
			Help: "Domain whois status, 0 means error, 1 means OK.",
		},
		[]string{"domain"},
	)

	DomainWhoisExpireDays = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_whois_expire_days",
			Help: "Domain whois expire days.",
		},
		[]string{"domain"},
	)
)

func init() {
	registry.MustRegister(DomainCertificateStatus)
	registry.MustRegister(DomainCertificateExpireDays)
	registry.MustRegister(DomainWhoisStatus)
	registry.MustRegister(DomainWhoisExpireDays)
}
