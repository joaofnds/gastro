package backuphealth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	syncTimeout   = 30 * time.Second
	pingTimeout   = 5 * time.Second
	responseLimit = 16 << 10
)

var (
	ErrConfig = errors.New("invalid backup health configuration")
	ErrSync   = errors.New("backup sync not confirmed")
	ErrPing   = errors.New("backup health ping failed")
)

type Config struct {
	URL string `mapstructure:"url"`
}

type Settings struct {
	DatabasePath string
	SocketPath   string
}

type Checker struct {
	pingURL      string
	databasePath string
	syncClient   *http.Client
	pingClient   *http.Client
}

func NewChecker(config Config, settings Settings, pingTransport *http.Transport) (*Checker, error) {
	if config.URL == "" {
		return &Checker{}, nil
	}

	pingURL, err := url.Parse(config.URL)
	if err != nil || pingURL.Scheme != "https" || pingURL.Hostname() == "" || pingURL.User != nil || pingURL.Fragment != "" {
		return nil, ErrConfig
	}

	if settings.DatabasePath == "" || !filepath.IsAbs(settings.SocketPath) || pingTransport == nil {
		return nil, ErrConfig
	}

	databasePath, err := filepath.Abs(settings.DatabasePath)
	if err != nil {
		return nil, ErrConfig
	}

	dialer := &net.Dialer{Timeout: syncTimeout}
	syncTransport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", settings.SocketPath)
		},
		MaxResponseHeaderBytes: responseLimit,
	}
	transport := pingTransport.Clone()
	transport.MaxResponseHeaderBytes = responseLimit
	transport.TLSHandshakeTimeout = pingTimeout

	return &Checker{
		pingURL:      pingURL.String(),
		databasePath: databasePath,
		syncClient:   &http.Client{Transport: syncTransport, Timeout: syncTimeout, CheckRedirect: rejectRedirect},
		pingClient:   &http.Client{Transport: transport, Timeout: pingTimeout, CheckRedirect: rejectRedirect},
	}, nil
}

func (checker *Checker) Check(ctx context.Context) error {
	if checker.pingURL == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, syncTimeout+pingTimeout)
	defer cancel()

	requestBody, err := json.Marshal(struct {
		Path    string `json:"path"`
		Wait    bool   `json:"wait"`
		Timeout int    `json:"timeout"`
	}{Path: checker.databasePath, Wait: true, Timeout: 30})
	if err != nil {
		return ErrSync
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/sync", bytes.NewReader(requestBody))
	if err != nil {
		return ErrSync
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := checker.syncClient.Do(request)
	if err != nil {
		return ErrSync
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return ErrSync
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, responseLimit+1))
	if err != nil || len(body) > responseLimit {
		return ErrSync
	}

	var syncResponse struct {
		Status         string  `json:"status"`
		Path           string  `json:"path"`
		TXID           *uint64 `json:"txid"`
		ReplicatedTXID *uint64 `json:"replicated_txid"`
	}
	if err := json.Unmarshal(body, &syncResponse); err != nil {
		return ErrSync
	}

	if (syncResponse.Status != "synced" && syncResponse.Status != "no_change") || syncResponse.Path != checker.databasePath {
		return ErrSync
	}

	if syncResponse.TXID == nil || syncResponse.ReplicatedTXID == nil || *syncResponse.TXID == 0 || *syncResponse.TXID != *syncResponse.ReplicatedTXID {
		return ErrSync
	}

	ping, err := http.NewRequestWithContext(ctx, http.MethodGet, checker.pingURL, nil)
	if err != nil {
		return ErrPing
	}

	result, err := checker.pingClient.Do(ping)
	if err != nil {
		return ErrPing
	}
	defer func() { _ = result.Body.Close() }()

	if result.StatusCode < http.StatusOK || result.StatusCode >= http.StatusMultipleChoices {
		return ErrPing
	}

	return nil
}

func (checker *Checker) Run(ctx context.Context, interval time.Duration, logger *zap.Logger) {
	if checker.pingURL == "" {
		return
	}

	defer checker.syncClient.CloseIdleConnections()
	defer checker.pingClient.CloseIdleConnections()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		if err := checker.Check(ctx); err != nil && ctx.Err() == nil {
			logger.Warn("backup health check failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func rejectRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}
