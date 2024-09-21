package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

func TokenBucketLimiting(genRate time.Duration, maxToken int) func(http.HandlerFunc) http.HandlerFunc {
	var tokens atomic.Int64
	tokens.Store(0)

	ticker := time.NewTicker(genRate)

	go func() {
		for {
			<-ticker.C
			if tokens.Load() < int64(maxToken) {
				tokens.Add(1)
			}
		}
	}()

	result := strconv.Itoa(int(genRate / time.Second))

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if tokens.Load() <= 0 {
				w.Header().Add("Retry-After", result)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			tokens.Add(-1)
			next(w, r)
		}
	}
}

func LeakyBucketLimiting(leakRate time.Duration, maxCapacity int) func(http.HandlerFunc) http.HandlerFunc {
	var bucket atomic.Int64
	bucket.Store(0)

	ticker := time.NewTicker(leakRate)

	// O balde vaza a uma taxa constante
	go func() {
		for {
			<-ticker.C
			if bucket.Load() > 0 {
				bucket.Add(-1)
			}
		}
	}()

	result := strconv.Itoa(int(leakRate / time.Second))

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if bucket.Load() >= int64(maxCapacity) {
				w.Header().Add("Retry-After", result)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			// Adiciona uma requisição ao balde
			bucket.Add(1)
			next(w, r)
		}
	}
}
