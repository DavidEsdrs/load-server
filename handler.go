package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Backend struct {
	target             *url.URL
	activesConnections uint64
	maxConn            uint64

	RateLimiter func(http.HandlerFunc) http.HandlerFunc

	accResponseTime  uint64
	totalConn        uint64
	meanResponseTime uint64
}

func NewBackend(target *url.URL) *Backend {
	return &Backend{
		target:             target,
		activesConnections: 0,
		maxConn:            1e3,
	}
}

func (b *Backend) SetMaxConn(maxConn uint64) {
	b.maxConn = maxConn
}

func (b *Backend) IncrementConn() {
	atomic.AddUint64(&b.activesConnections, 1)
}

func (b *Backend) DecrementConn() {
	atomic.AddUint64(&b.activesConnections, ^uint64(0))
}

func (b *Backend) IncrementTotal() {
	atomic.AddUint64(&b.totalConn, 1)
}

func (b *Backend) Cleanup() error {
	return nil
}

type ProxyHandler struct {
	backends    []*Backend
	currBackend uint64
	maxBackend  int
}

func (ph *ProxyHandler) Cleanup() error {
	for _, b := range ph.backends {
		if err := b.Cleanup(); err != nil {
			return err
		}
	}
	return nil
}

// create a ProxyHandler that can redirect requests to given backends
func NewProxyHandler(backends []BackendConfig) *ProxyHandler {
	var backendURLs []*Backend

	for _, b := range backends {
		url, err := url.Parse(b.Path)
		if err != nil {
			panic("Invalid backend path")
		}
		backend := NewBackend(url)

		if limiter, ok := b.RateLimit.(map[string]interface{}); ok {
			switch limiter["type"] {
			case "token_bucket":
				backend.RateLimiter = TokenBucketLimiting(
					time.Duration(limiter["generation_time_ms"].(int))*time.Millisecond,
					limiter["max_token"].(int),
				)
			case "leaky_bucket":
				backend.RateLimiter = LeakyBucketLimiting(
					time.Duration(limiter["leaky_rate_ms"].(int))*time.Millisecond,
					limiter["max_capacity"].(int),
				)
			default:
				backend.RateLimiter = func(hf http.HandlerFunc) http.HandlerFunc {
					return hf
				}
			}

			log.Println("backend", b.Path, "added -", limiter["type"], "rate limit set")
		} else {
			backend.RateLimiter = func(hf http.HandlerFunc) http.HandlerFunc {
				return hf
			}
			log.Println("backend", b.Path, "added -", "no rate limit set")
		}

		backendURLs = append(backendURLs, backend)
	}

	return &ProxyHandler{
		backends:    backendURLs,
		currBackend: 0,
		maxBackend:  len(backendURLs),
	}
}

func (ph *ProxyHandler) RoundRobin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next := atomic.AddUint64(&ph.currBackend, 1)
		targetBackend := ph.backends[next%uint64(ph.maxBackend)]
		proxy := httputil.NewSingleHostReverseProxy(targetBackend.target)
		targetBackend.RateLimiter(proxy.ServeHTTP)(w, r)
	}
}

func (ph *ProxyHandler) Random() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next := rand.Intn(ph.maxBackend)
		targetBackend := ph.backends[next]
		proxy := httputil.NewSingleHostReverseProxy(targetBackend.target)
		targetBackend.RateLimiter(proxy.ServeHTTP)(w, r)
	}
}

func (ph *ProxyHandler) LeastConnection() http.HandlerFunc {
	var mu sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {
		targetBackend := ph.backends[0]

		for _, b := range ph.backends {
			if b.activesConnections < targetBackend.activesConnections {
				targetBackend = b
			}
		}

		requestId := GenerateRequestID()
		proxy := httputil.NewSingleHostReverseProxy(targetBackend.target)
		originalDirector := proxy.Director

		proxy.Director = func(r *http.Request) {
			mu.Lock()
			defer mu.Unlock()

			originalDirector(r)
			r.Header.Set("X-Request-ID", requestId)
			targetBackend.IncrementConn()
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			mu.Lock()
			defer mu.Unlock()

			r.Header.Set("X-Request-ID", requestId)
			targetBackend.DecrementConn()
			return nil
		}

		targetBackend.RateLimiter(proxy.ServeHTTP)(w, r)
	}
}

func (ph *ProxyHandler) LeastResponseTime() http.HandlerFunc {
	var mu sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {
		targetBackend := ph.backends[0]

		for _, b := range ph.backends {
			if b.meanResponseTime < targetBackend.meanResponseTime {
				targetBackend = b
			}
		}

		requestId := GenerateRequestID()
		proxy := httputil.NewSingleHostReverseProxy(targetBackend.target)
		originalDirector := proxy.Director

		var (
			start    time.Time
			duration time.Duration
		)

		proxy.Director = func(r *http.Request) {
			mu.Lock()
			defer mu.Unlock()

			originalDirector(r)
			r.Header.Set("X-Request-ID", requestId)
			targetBackend.IncrementConn()
			targetBackend.IncrementTotal()
			start = time.Now()
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			mu.Lock()
			defer mu.Unlock()

			r.Header.Set("X-Request-ID", requestId)
			targetBackend.DecrementConn()
			duration = time.Since(start)
			calculateMeanTime(targetBackend, uint64(duration.Milliseconds()))
			return nil
		}

		targetBackend.RateLimiter(proxy.ServeHTTP)(w, r)
	}
}

func (ph *ProxyHandler) WithRateLimiting(
	rateLimiter func(http.HandlerFunc) http.HandlerFunc,
	balancer http.HandlerFunc,
) http.HandlerFunc {
	return rateLimiter(balancer)
}

func GenerateRequestID() string {
	return uuid.New().String()
}

func calculateMeanTime(backend *Backend, duration uint64) {
	backend.totalConn++
	backend.accResponseTime += duration
	backend.meanResponseTime = backend.accResponseTime / backend.totalConn
}
