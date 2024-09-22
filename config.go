package main

import (
	"log"
	"net/http"
)

type TokenBucket struct {
	Type           string `mapstructure:"type"`
	GenerationTime int    `mapstructure:"generation_time_ms"`
	MaxToken       int    `mapstructure:"max_token"`
}

type LeakyBucket struct {
	Type        string `mapstructure:"type"`
	LeakyRateMs int    `mapstructure:"leaky_rate_ms"`
	MaxCapacity int    `mapstructure:"max_capacity"`
}

type BackendConfig struct {
	Name      string      `mapstructure:"name"`
	Path      string      `mapstructure:"path"`
	RateLimit interface{} `mapstructure:"rate_limit"`
}

type RoutingRules struct {
	Path    string `mapstructure:"path"`
	Backend string `mapstructure:"backend"`
}

type Config struct {
	Port      string `mapstructure:"port"`
	Balancing struct {
		Type string `mapstructure:"type"`
	} `mapstructure:"balancing"`
	BackendConfigs []BackendConfig `mapstructure:"backends"`
	RoutingRules   []RoutingRules  `mapstructure:"routing_rules"`
}

func getBalancingAlgorithm(proxyHandler *ProxyHandler) func() http.HandlerFunc {
	var (
		balancingAlgorithm func() http.HandlerFunc
	)

	switch config.Balancing.Type {
	case "round_robin":
		balancingAlgorithm = proxyHandler.RoundRobin
	case "random":
		balancingAlgorithm = proxyHandler.Random
	case "least_connections":
		balancingAlgorithm = proxyHandler.LeastConnection
	case "least_response_time":
		balancingAlgorithm = proxyHandler.LeastResponseTime
	default:
		log.Fatal("unknown balancing algorithm")
	}

	return balancingAlgorithm
}
