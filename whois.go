package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func GetWhoisTimeout(domain string, timeout time.Duration) (string, error) {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		err := fmt.Errorf("Domain(%s) name is wrong!", domain)
		return "", err
	}
	//last part of domain is zome
	zone := parts[len(parts)-1]
	defaultServer := fmt.Sprintf("whois.nic.%s", zone)
	server, ok := servers[zone]
	if !ok {
		server = defaultServer
	}
	ret, err := GetWhoisWithServerTimeout(domain, server, timeout)
	if err != nil {
		return ret, err
	}
	if ret == "" && server != defaultServer {
		return GetWhoisWithServerTimeout(domain, defaultServer, timeout)
	}
	return ret, err
}

func GetWhoisWithServerTimeout(domain, server string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(server, "43"), timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.Write([]byte(domain + "\r\n"))
	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

type WhoisResult struct {
	Domain     string
	Status     string
	ErrorMsg   string
	ExpireAt   time.Time
	ExpireDays int
}

type WhoisResults map[string]WhoisResult

type WhoisChecker struct {
	Domains []string
}

func NewWhoisChecker(domains []string) *WhoisChecker {
	return &WhoisChecker{
		Domains: domains,
	}
}

func (wc *WhoisChecker) Check() WhoisResults {
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
		ret  WhoisResults = make(WhoisResults)
	)
	wg.Add(len(wc.Domains))
	for _, item := range wc.Domains {
		go func(domain string) {
			cr := wc.CheckOneDomain(domain)
			lock.Lock()
			ret[domain] = cr
			lock.Unlock()
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ret
}

func (wc *WhoisChecker) CheckOneDomain(domain string) WhoisResult {
	ret := WhoisResult{
		Domain: domain,
		Status: "Error",
	}
	whois, err := GetWhoisTimeout(domain, 5*time.Second)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%s", err)
		return ret
	}
	et, days := wc.decodeWhoisInfo(whois)
	log.Println("[INFO] Whois", domain, "Expire After", days, "Days,", et)
	ret.Status = "OK"
	ret.ExpireAt = et
	ret.ExpireDays = days
	return ret
}

func (wc *WhoisChecker) decodeWhoisInfo(info string) (time.Time, int) {
	for _, rline := range strings.Split(info, "\n") {
		if strings.Contains(rline, "Expir") {
			line := strings.TrimSpace(rline)
			parts := strings.Split(line, ": ")
			if len(parts) == 2 {
				expireDate := strings.TrimSpace(parts[1])
				expireDate = strings.ReplaceAll(expireDate, "T", " ")
				eparts := strings.Split(strings.ToUpper(expireDate), "Z")
				et := parseDateStr(eparts[0])
				days := int(et.Sub(time.Now()).Hours() / 24)
				return et, days
			}
		} else if strings.Contains(rline, "有効期限") {
			line := strings.TrimSpace(rline)
			parts := strings.Split(line, "]")
			if len(parts) == 2 {
				dateStr := strings.TrimSpace(parts[1])
				et := parseJPDateStr(dateStr)
				days := int(et.Sub(time.Now()).Hours() / 24)
				return et, days
			}
		}
	}
	log.Println("----Error Cannot Parse Whois Info----")
	log.Println(info)
	log.Println("-------------------------------------")
	return time.Time{}, 0
}

func parseDateStr(date string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		log.Println("[Error]", err)
		return time.Time{}
	}
	return t
}

func parseJPDateStr(date string) time.Time {
	parts := strings.Split(date, "/")
	if len(parts) != 3 {
		return time.Time{}
	}
	year, month, day := parts[0], parts[1], parts[2]
	t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", year, month, day))
	if err != nil {
		log.Println("[Error]", err)
		return time.Time{}
	}
	return t
}
