package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DeviceMetrics struct {
	Power           float64 `json:"power"`
	Ws              float64 `json:"Ws"`
	Relay           bool    `json:"relay"`
	Temperature     float64 `json:"temperature"`
	EnergySinceBoot float64 `json:"energy_since_boot"`
	TimeSinceBoot   int64   `json:"time_since_boot"`
}

var (
	power = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_power_watts",
		Help: "Power consumption in watts",
	})
	Ws = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_energy_ws",
		Help: "Energy usage in watt-seconds",
	})
	relay = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_relay_status",
		Help: "Status of the relay (0: off, 1: on)",
	})
	temperature = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_temperature_celsius",
		Help: "Temperature in Celsius",
	})
	energySinceBoot = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_energy_since_boot",
		Help: "Energy consumed since last boot",
	})
	timeSinceBoot = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_time_since_boot",
		Help: "Time elapsed since last boot",
	})
)

func init() {
	// Register metrics.
	prometheus.MustRegister(power, Ws, relay, temperature, energySinceBoot, timeSinceBoot)
}

func fetchMetrics() {
	err := godotenv.Load() // Loads the .env file that is in the same directory as the code
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ip := os.Getenv("IP_WIFI_SWITCH")
	url := fmt.Sprintf("http://%s/report", ip)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching metrics: %v", err)
		return
	}
	defer resp.Body.Close()

	var metrics DeviceMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	power.Set(metrics.Power)
	Ws.Set(metrics.Ws)
	if metrics.Relay {
		relay.Set(1)
	} else {
		relay.Set(0)
	}
	temperature.Set(metrics.Temperature)
	energySinceBoot.Set(metrics.EnergySinceBoot)
	timeSinceBoot.Set(float64(metrics.TimeSinceBoot))
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		for {
			fetchMetrics()
			// Fetch metrics every 10 seconds
			time.Sleep(10 * time.Second)
		}
	}()

	log.Fatal(http.ListenAndServe(":5000", nil))
}
