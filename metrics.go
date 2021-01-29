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

	DomainResolveStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_resolve_status",
			Help: "Domain resolve status, 0 means error, 1 means OK.",
		},
		[]string{"domain"},
	)

	DomainResolveIPs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_resolve_ips",
			Help: "Domain resolved IP addresses",
		},
		[]string{"domain"},
	)

	DomainRequestStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_request_status",
			Help: "Domain request status, 0 means error, 1 means OK.",
		},
		[]string{"domain", "host", "path"},
	)

	DomainRequestError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "domain_request_error",
			Help: "Domain request error Log.",
		},
		[]string{"domain", "host", "path", "address", "status"},
	)
)

func init() {
	registry.MustRegister(DomainCertificateStatus)
	registry.MustRegister(DomainCertificateExpireDays)
	registry.MustRegister(DomainWhoisStatus)
	registry.MustRegister(DomainWhoisExpireDays)
	registry.MustRegister(DomainResolveStatus)
	registry.MustRegister(DomainResolveIPs)
	registry.MustRegister(DomainRequestStatus)
	registry.MustRegister(DomainRequestError)
}
