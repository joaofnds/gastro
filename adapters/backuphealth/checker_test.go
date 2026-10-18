package backuphealth_test

import (
	"astro/adapters/backuphealth"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestBackupHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backup health")
}

var _ = Describe("Checker", func() {
	It("pings after a waited sync confirms the configured database reached remote storage", func() {
		requests := make(chan syncRequest, 1)
		socket := serveSync(func(w http.ResponseWriter, r *http.Request) {
			var request syncRequest
			_ = json.NewDecoder(r.Body).Decode(&request)
			request.Method = r.Method
			request.Endpoint = r.URL.Path
			request.ContentType = r.Header.Get("Content-Type")
			requests <- request
			_, _ = w.Write([]byte(`{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":42}`))
		})
		pings := make(chan string, 1)
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pings <- r.Method + " " + r.URL.RequestURI()
			w.WriteHeader(http.StatusOK)
		}))
		DeferCleanup(receiver.Close)
		checker, err := backuphealth.NewChecker(backuphealth.Config{URL: receiver.URL + "/secret"}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: socket,
		}, trustReceiver(receiver))
		Expect(err).NotTo(HaveOccurred())

		err = checker.Check(GinkgoT().Context())

		Expect(err).NotTo(HaveOccurred())
		Expect(requests).To(Receive(Equal(syncRequest{
			Method: http.MethodPost, Endpoint: "/sync", ContentType: "application/json",
			Path: "/data/astro.db", Wait: true, Timeout: 30,
		})))
		Expect(pings).To(Receive(Equal("GET /secret")))
	})

	It("accepts an idle sync with an established matching remote position", func() {
		socket := serveSync(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"status":"no_change","path":"/data/astro.db","txid":42,"replicated_txid":42}`))
		})
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		DeferCleanup(receiver.Close)
		checker := newChecker(socket, receiver)

		Expect(checker.Check(GinkgoT().Context())).To(Succeed())
	})

	It("withholds the ping while remote sync is pending", func() {
		started := make(chan struct{})
		release := make(chan struct{})
		socket := serveSync(func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			<-release
			_, _ = w.Write([]byte(`{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":42}`))
		})
		DeferCleanup(func() { close(release) })
		var pings atomic.Int32
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			pings.Add(1)
		}))
		DeferCleanup(receiver.Close)
		checker := newChecker(socket, receiver)
		result := make(chan error, 1)

		go func() { result <- checker.Check(GinkgoT().Context()) }()
		Eventually(started).Should(BeClosed())

		Consistently(pings.Load, 50*time.Millisecond).Should(BeZero())
		release <- struct{}{}
		Eventually(result).Should(Receive(Succeed()))
		Expect(pings.Load()).To(Equal(int32(1)))
	})

	DescribeTable("withholds success for an unconfirmed sync", func(status int, body string) {
		socket := serveSync(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		})
		var pings atomic.Int32
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			pings.Add(1)
		}))
		DeferCleanup(receiver.Close)
		checker := newChecker(socket, receiver)

		err := checker.Check(GinkgoT().Context())

		Expect(err).To(MatchError(backuphealth.ErrSync))
		Expect(pings.Load()).To(BeZero())
	},
		Entry("upload failure", http.StatusInternalServerError, `{"error":"https://secret.invalid/token"}`),
		Entry("partial backlog", http.StatusOK, `{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":41}`),
		Entry("local completion", http.StatusOK, `{"status":"synced_local","path":"/data/astro.db","txid":42,"replicated_txid":42}`),
		Entry("different database", http.StatusOK, `{"status":"synced","path":"/data/other.db","txid":42,"replicated_txid":42}`),
		Entry("missing remote position", http.StatusOK, `{"status":"synced","path":"/data/astro.db","txid":42}`),
		Entry("missing local position", http.StatusOK, `{"status":"synced","path":"/data/astro.db","replicated_txid":42}`),
		Entry("uninitialized database", http.StatusOK, `{"status":"no_change","path":"/data/astro.db","txid":0,"replicated_txid":0}`),
		Entry("invalid position", http.StatusOK, `{"status":"synced","path":"/data/astro.db","txid":"42","replicated_txid":42}`),
		Entry("negative position", http.StatusOK, `{"status":"synced","path":"/data/astro.db","txid":-1,"replicated_txid":-1}`),
		Entry("malformed JSON", http.StatusOK, `{"status":`),
		Entry("trailing JSON", http.StatusOK, `{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":42}{}`),
		Entry("oversized response", http.StatusOK, strings.Repeat(" ", 16<<10)+`{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":42}`),
	)

	It("withholds success when the control socket is unavailable", func() {
		checker, err := backuphealth.NewChecker(backuphealth.Config{URL: "https://health.invalid/secret"}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: filepath.Join(GinkgoT().TempDir(), "absent.sock"),
		}, &http.Transport{})
		Expect(err).NotTo(HaveOccurred())

		Expect(checker.Check(GinkgoT().Context())).To(MatchError(backuphealth.ErrSync))
	})

	It("stays disabled without a URL", func() {
		checker, err := backuphealth.NewChecker(backuphealth.Config{}, backuphealth.Settings{}, nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(checker.Check(GinkgoT().Context())).To(Succeed())
		checker.Run(GinkgoT().Context(), time.Minute, zap.NewNop())
	})

	DescribeTable("rejects invalid ping URLs without returning their contents", func(pingURL string) {
		_, err := backuphealth.NewChecker(backuphealth.Config{URL: pingURL}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: "/tmp/litestream.sock",
		}, &http.Transport{})

		Expect(err).To(MatchError(backuphealth.ErrConfig))
	},
		Entry("plain HTTP", "http://health.invalid/secret"),
		Entry("relative URL", "/secret"),
		Entry("missing host", "https:///secret"),
		Entry("credentials", "https://user:password@health.invalid/secret"),
		Entry("fragment", "https://health.invalid/secret#token"),
		Entry("invalid port", "https://health.invalid:token/secret"),
		Entry("malformed escape", "https://health.invalid/%token"),
	)

	It("rejects an enabled checker without its HTTP transport", func() {
		_, err := backuphealth.NewChecker(backuphealth.Config{URL: "https://health.invalid/secret"}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: "/tmp/litestream.sock",
		}, nil)

		Expect(err).To(MatchError(backuphealth.ErrConfig))
	})

	It("rejects ping redirects without contacting their destination", func() {
		socket := completedSync()
		var redirected atomic.Int32
		destination := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			redirected.Add(1)
		}))
		DeferCleanup(destination.Close)
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, destination.URL+"/secret", http.StatusFound)
		}))
		DeferCleanup(receiver.Close)
		checker := newChecker(socket, receiver)

		err := checker.Check(GinkgoT().Context())

		Expect(err).To(MatchError(backuphealth.ErrPing))
		Expect(redirected.Load()).To(BeZero())
	})

	It("returns a safe error when the ping receiver fails", func() {
		socket := completedSync()
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "https://health.invalid/secret", http.StatusInternalServerError)
		}))
		DeferCleanup(receiver.Close)
		checker := newChecker(socket, receiver)

		Expect(checker.Check(GinkgoT().Context())).To(MatchError(backuphealth.ErrPing))
	})

	It("logs only the safe error when an HTTP client failure includes the secret URL", func() {
		socket := completedSync()
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		transport := trustReceiver(receiver)
		pingURL := receiver.URL + "/private-token?secret=value"
		receiver.Close()
		checker, err := backuphealth.NewChecker(backuphealth.Config{URL: pingURL}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: socket,
		}, transport)
		Expect(err).NotTo(HaveOccurred())
		core, logs := observer.New(zap.WarnLevel)
		ctx, cancel := context.WithCancel(GinkgoT().Context())
		defer cancel()
		done := make(chan struct{})

		go func() {
			defer close(done)
			checker.Run(ctx, time.Minute, zap.New(core))
		}()
		Eventually(logs.Len).Should(Equal(1))
		cancel()
		Eventually(done).Should(BeClosed())

		record := logs.All()[0]
		Expect(record.Message).To(Equal("backup health check failed"))
		Expect(record.ContextMap()).To(Equal(map[string]any{"error": backuphealth.ErrPing.Error()}))
	})

	It("bounds a stalled sync response by the caller deadline", func() {
		started := make(chan struct{})
		socket := serveSync(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			w.WriteHeader(http.StatusOK)
			_ = http.NewResponseController(w).Flush()
			close(started)
			<-r.Context().Done()
		})
		checker, err := backuphealth.NewChecker(backuphealth.Config{URL: "https://health.invalid/secret"}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: socket,
		}, &http.Transport{})
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := context.WithTimeout(GinkgoT().Context(), 100*time.Millisecond)
		defer cancel()

		err = checker.Check(ctx)

		Expect(started).To(BeClosed())
		Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
		Expect(err).To(MatchError(backuphealth.ErrSync))
	})

	It("keeps slow checks sequential and cancels the pending request", func() {
		var active, maximum atomic.Int32
		started := make(chan struct{}, 10)
		socket := serveSync(func(_ http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			current := active.Add(1)
			maximum.Store(max(maximum.Load(), current))
			defer active.Add(-1)
			started <- struct{}{}
			<-r.Context().Done()
		})
		checker, err := backuphealth.NewChecker(backuphealth.Config{URL: "https://health.invalid/secret"}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: socket,
		}, &http.Transport{})
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := context.WithTimeout(GinkgoT().Context(), 100*time.Millisecond)
		defer cancel()
		done := make(chan struct{})

		go func() {
			defer close(done)
			checker.Run(ctx, time.Millisecond, zap.NewNop())
		}()
		Eventually(started).Should(Receive())
		Eventually(done).Should(BeClosed())

		Expect(maximum.Load()).To(Equal(int32(1)))
		Expect(started).NotTo(Receive())
		Eventually(active.Load).Should(BeZero())
	})

	It("repeats completed sync and ping cycles until canceled", func() {
		syncs := make(chan struct{}, 8)
		socket := serveSync(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":42}`))
			select {
			case syncs <- struct{}{}:
			case <-r.Context().Done():
			}
		})
		pings := make(chan struct{}, 8)
		receiver := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			select {
			case pings <- struct{}{}:
			case <-r.Context().Done():
			}
		}))
		DeferCleanup(receiver.Close)
		checker := newChecker(socket, receiver)
		ctx, cancel := context.WithCancel(GinkgoT().Context())
		done := make(chan struct{})
		DeferCleanup(func() {
			cancel()
			Eventually(done).Should(BeClosed())
		})

		go func() {
			defer close(done)
			checker.Run(ctx, 10*time.Millisecond, zap.NewNop())
		}()

		Eventually(syncs).Should(Receive())
		Eventually(pings).Should(Receive())
		Eventually(syncs).Should(Receive())
		Eventually(pings).Should(Receive())
		cancel()
		Eventually(done).Should(BeClosed())
	})

	It("cancels its pending check when the Fx module stops", func() {
		started := make(chan struct{})
		canceled := make(chan struct{})
		socket := serveSync(func(_ http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			close(started)
			<-r.Context().Done()
			close(canceled)
		})
		checker, err := backuphealth.NewChecker(backuphealth.Config{URL: "https://health.invalid/secret"}, backuphealth.Settings{
			DatabasePath: "/data/astro.db", SocketPath: socket,
		}, &http.Transport{})
		Expect(err).NotTo(HaveOccurred())
		app := fxtest.New(GinkgoT(), backuphealth.Module, fx.Replace(checker), fx.Supply(zap.NewNop()))
		DeferCleanup(app.RequireStop)
		app.RequireStart()
		Eventually(started).Should(BeClosed())

		app.RequireStop()

		Eventually(canceled).Should(BeClosed())
	})
})

type syncRequest struct {
	Method      string
	Endpoint    string
	ContentType string
	Path        string `json:"path"`
	Wait        bool   `json:"wait"`
	Timeout     int    `json:"timeout"`
}

func serveSync(handler http.HandlerFunc) string {
	socket := filepath.Join(GinkgoT().TempDir(), "litestream.sock")
	listener, err := net.Listen("unix", socket)
	Expect(err).NotTo(HaveOccurred())
	server := httptest.NewUnstartedServer(handler)
	_ = server.Listener.Close()
	server.Listener = listener
	server.Start()
	DeferCleanup(func() {
		server.CloseClientConnections()
		server.Close()
	})
	return socket
}

func trustReceiver(receiver *httptest.Server) *http.Transport {
	roots := x509.NewCertPool()
	roots.AddCert(receiver.Certificate())
	return &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
}

func newChecker(socket string, receiver *httptest.Server) *backuphealth.Checker {
	checker, err := backuphealth.NewChecker(backuphealth.Config{URL: receiver.URL + "/secret"}, backuphealth.Settings{
		DatabasePath: "/data/astro.db", SocketPath: socket,
	}, trustReceiver(receiver))
	Expect(err).NotTo(HaveOccurred())
	return checker
}

func completedSync() string {
	return serveSync(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"synced","path":"/data/astro.db","txid":42,"replicated_txid":42}`))
	})
}
