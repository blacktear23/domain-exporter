# Domain Exporter

Domain Exporter is a prometheus exporter for check domain expiration and HTTPS certification expiration.

# Usage

```
./domain-exporter -config ./config.yaml -web.listen-address :9170 -web.telemetry-path /metrics
```

After you update the `config.yaml` you can use `kill -HUP` to let domain-exporter to reload.

# Configuration

```yaml
# Collect duration
collect_duration: 3600

# Certificate check domains
certificate_domains:
  - www.baidu.com

# Whois check domains
whois_domains:
  - baidu.com
```

* collect\_duration: Collector duration, unit is second. This will control how long the Collector recheck the domains.
* certificate\_domains: HTTPS domains that need to be checked
* whois\_domains: Whois domains that need to be checked

# Metrics

Example:

```
# HELP domain_certificate_expire_days Domain certificate expire days.
# TYPE domain_certificate_expire_days gauge
domain_certificate_expire_days{domain="www.baidu.com"} 348
# HELP domain_certificate_status Domain certificate status, 0 means error, 1 means OK.
# TYPE domain_certificate_status gauge
domain_certificate_status{domain="www.baidu.com"} 1
# HELP domain_whois_expire_days Domain whois expire days.
# TYPE domain_whois_expire_days gauge
domain_whois_expire_days{domain="baidu.com"} 2251
# HELP domain_whois_status Domain whois status, 0 means error, 1 means OK.
# TYPE domain_whois_status gauge
domain_whois_status{domain="baidu.com"} 1
```
