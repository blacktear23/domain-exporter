package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const VERSION = "1.0"

func main() {
	var (
		listenAddr  string
		metricsPath string
		configFile  string
	)

	flag.StringVar(&listenAddr, "web.listen-address", ":9170", "An address to listen for web interface and telemetry.")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.Parse()
	log.SetOutput(os.Stdout)

	cfg, err := NewConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	collector := NewCollector(cfg)
	dnsProber := NewDNSProber(cfg)

	log.Printf("Start Domain Checker Prometheus Exporter Version=%v", VERSION)
	log.Printf("Collect Duration: %v", cfg.GetDuration())

	// Start collector
	go collector.Start()

	// Handle for metrics path
	http.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Handle for / path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, `<!DOCTYPE html>
			<title>Domain Exporter</title>
			<h1>Domain Exporter</h1>
			<p><a href=%q>Metrics</a></p>`,
			metricsPath)
		if err != nil {
			log.Printf("Error while sending a response for '/' path: %v", err)
		}
	})

	// Handle for /probe path
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		result, err := dnsProber.Probe()
		if err != nil {
			log.Printf("Error while generate response for '/probe' path: %v", err)
		}
		_, err = w.Write(result)
		if err != nil {
			log.Printf("Error while sending a response for '/probe' path: %v", err)
		}
	})

	log.Printf("Start Web Server At: %s", listenAddr)
	// Start Web Server
	go func() {
		log.Fatal(http.ListenAndServe(listenAddr, nil))
	}()

	waitSignal(func() {
		err := cfg.Reload()
		if err != nil {
			log.Println(err)
		} else {
			ResetAllMetrics()
			collector.CollectOnce()
		}
	}, nil)
}

type SignalHandler func()

func waitSignal(onReload, onExit SignalHandler) {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGHUP)
	for sig := range sigChan {
		if sig == syscall.SIGHUP {
			if onReload != nil {
				log.Println("Reloading")
				onReload()
			}
		} else {
			if onExit != nil {
				onExit()
			}
			log.Fatal("Server Exit\n")
		}
	}
}
