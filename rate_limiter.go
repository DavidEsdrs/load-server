package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

func LeakyBucket(genRate time.Duration, maxToken int) func(http.HandlerFunc) http.HandlerFunc {
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
