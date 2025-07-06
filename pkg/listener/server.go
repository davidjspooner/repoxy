package listener

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Config struct {
	ExternalUrl  string   `yaml:"url"`
	InternalPort int      `yaml:"port"`
	Features     []string `yaml:"features"`
	InternalAddr string   `yaml:"address"` // optional, defaults to all interfaces
	Tls          struct {
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls"`
	certCache     *tls.Certificate
	certCacheTime time.Time
	certMutex     sync.RWMutex
}

func (l *Config) Errorf(format string, args ...interface{}) error {
	return fmt.Errorf("listener %s:%d: "+format, append([]interface{}{l.InternalAddr, l.InternalPort}, args...)...)
}

var ctxKey = &struct{}{}

func (l *Config) FromRequest(r *http.Request) *Config {
	if r == nil {
		return nil
	}
	listener, ok := r.Context().Value(ctxKey).(*Config)
	if !ok || listener == nil {
		return nil
	}
	return listener
}

func (l *Config) server(handler http.Handler) (*http.Server, error) {
	if l.InternalPort <= 0 || l.InternalPort > 65535 {
		return nil, l.Errorf("invalid port %d", l.InternalPort)
	}
	if l.ExternalUrl == "" {
		return nil, l.Errorf("external URL is required")
	}
	s := &http.Server{
		Addr: fmt.Sprintf("%s:%d", l.InternalAddr, l.InternalPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ctxKey, l))
			handler.ServeHTTP(w, r)
		}),
	}
	if l.Tls.CertFile != "" || l.Tls.KeyFile != "" {
		if l.Tls.CertFile == "" || l.Tls.KeyFile == "" {
			return nil, l.Errorf("both cert_file and key_file must be provided for TLS")
		}
		s.TLSConfig = &tls.Config{
			GetCertificate: l.getCertificate,
		}
	}
	return s, nil
}

func (l *Config) getCachedCertificate(hello *tls.ClientHelloInfo, now time.Time) (*tls.Certificate, error) {
	l.certMutex.RLock()
	defer l.certMutex.RUnlock()

	// Return cached certificate if it exists and is fresh
	if l.certCache != nil {
		age := now.Sub(l.certCacheTime)
		if age < 5*time.Minute {
			return l.certCache, nil
		}

		// Cache is old, check if file has been modified
		stat, err := os.Stat(l.Tls.CertFile)
		if err != nil {
			return nil, l.Errorf("failed to stat TLS cert file %s: %v", l.Tls.CertFile, err)
		}

		// If file hasn't been modified since we cached it, return cached cert
		if !stat.ModTime().After(l.certCacheTime) {
			l.certCacheTime = now // Update cache time to now
			return l.certCache, nil
		}
	}

	return nil, nil
}

func (l *Config) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if l.Tls.CertFile == "" || l.Tls.KeyFile == "" {
		return nil, l.Errorf("TLS certificate files are not configured")
	}

	now := time.Now()
	cached, err := l.getCachedCertificate(hello, now)
	if cached != nil || err != nil {
		return cached, err
	}

	l.certMutex.Lock()
	defer l.certMutex.Unlock()

	// Load certificate (first time or file was modified)
	cert, err := tls.LoadX509KeyPair(l.Tls.CertFile, l.Tls.KeyFile)
	if err != nil {
		return nil, l.Errorf("failed to load TLS certificate: %v", err)
	}

	l.certCache = &cert
	l.certCacheTime = now
	return l.certCache, nil
}
