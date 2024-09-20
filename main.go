package main

import (
	"bufio"
	"net/http"
	"os"
	"time"
)

func testing() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 200)
		w.Write([]byte("chegou em 3001"))
	})
	go http.ListenAndServe(":3001", mux)

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 300)
		w.Write([]byte("chegou em 3002"))
	})
	go http.ListenAndServe(":3002", mux2)

	mux3 := http.NewServeMux()
	mux3.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 100)
		w.Write([]byte("chegou em 3003"))
	})
	go http.ListenAndServe(":3003", mux3)
}

func main() {
	testing()

	backends := parseBackends("backends")

	proxyHandler := NewProxyHandler(backends)

	http.ListenAndServe(
		":3000",
		proxyHandler.LeastConnection(),
	)
}

func parseBackends(filename string) []string {
	var backends []string

	f, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		backends = append(backends, scanner.Text())
	}

	return backends
}
