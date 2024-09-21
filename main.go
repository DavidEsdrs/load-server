package main

import (
	"log"
	"net/http"
)

var config Config

func testing() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chegou em 3001"))
	})
	go http.ListenAndServe(":3001", mux)

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chegou em 3002"))
	})
	go http.ListenAndServe(":3002", mux2)

	mux3 := http.NewServeMux()
	mux3.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chegou em 3003"))
	})
	go http.ListenAndServe(":3003", mux3)
}

func main() {
	testing()

	parseConfig("balance")

	proxyHandler := NewProxyHandler(config.BackendConfigs)
	defer proxyHandler.Cleanup()

	balancingAlgorithm := getBalancingAlgorithm(proxyHandler)

	log.Fatal(http.ListenAndServe(config.Port, balancingAlgorithm()))
}
