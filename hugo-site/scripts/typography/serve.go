// This file proxies Hugo preview traffic so generated HTML uses the same transform as builds.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func serveHugo(arguments []string) error {
	options := flag.NewFlagSet("serve", flag.ContinueOnError)
	bind := options.String("bind", "127.0.0.1", "preview listener address")
	baseURL := options.String("baseURL", "http://localhost:1313/", "browser-facing preview URL")
	if err := options.Parse(arguments); err != nil {
		return err
	}
	if options.NArg() != 0 {
		return fmt.Errorf("unexpected serve arguments: %s", strings.Join(options.Args(), " "))
	}
	listenAddress := net.JoinHostPort(*bind, "1313")
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", listenAddress, err)
	}
	upstream, err := url.Parse("http://127.0.0.1:1315")
	if err != nil {
		return errors.Join(fmt.Errorf("parse Hugo upstream URL: %w", err), listener.Close())
	}
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return errors.Join(errors.New("default HTTP transport is not cloneable"), listener.Close())
	}
	transport := defaultTransport.Clone()
	defer transport.CloseIdleConnections()
	transport.DisableCompression = true
	proxy.Transport = transport
	proxy.Director = func(request *http.Request) {
		request.URL.Scheme = upstream.Scheme
		request.URL.Host = upstream.Host
		request.Host = upstream.Host
		request.Header.Del("Accept-Encoding")
	}
	proxy.ModifyResponse = transformResponse
	server := &http.Server{Handler: proxy, ReadHeaderTimeout: 10 * time.Second}
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()

	command := exec.Command("hugo", "server", "--bind", "127.0.0.1", "--port", "1315", "--liveReloadPort", "1313", "--disableFastRender", "--renderToMemory", "--baseURL", *baseURL, "--appendPort=false")
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Start(); err != nil {
		return errors.Join(fmt.Errorf("start Hugo server: %w", err), server.Close())
	}
	childResult := make(chan error, 1)
	go func() { childResult <- command.Wait() }()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-childResult:
		closeErr := server.Close()
		if err == nil {
			return errors.Join(errors.New("Hugo server exited unexpectedly"), closeErr)
		}
		return errors.Join(fmt.Errorf("Hugo server exited: %w", err), closeErr)
	case err := <-serveResult:
		childErr := stopChild(command, childResult)
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.Join(fmt.Errorf("preview proxy failed: %w", err), childErr)
		}
		return childErr
	case <-signals:
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownContext)
		childErr := stopChild(command, childResult)
		if shutdownErr != nil {
			return fmt.Errorf("shut down preview proxy: %w", shutdownErr)
		}
		if childErr != nil {
			return fmt.Errorf("stop Hugo server: %w", childErr)
		}
		return nil
	}
}

func stopChild(command *exec.Cmd, childResult <-chan error) error {
	if command.Process == nil {
		return nil
	}
	if err := command.Process.Signal(os.Interrupt); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	select {
	case err := <-childResult:
		if err != nil && wasSignaled(command) {
			return nil
		}
		return err
	case <-time.After(10 * time.Second):
		if err := command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		return <-childResult
	}
}

func wasSignaled(command *exec.Cmd) bool {
	if command.ProcessState == nil {
		return false
	}
	waitStatus, ok := command.ProcessState.Sys().(syscall.WaitStatus)
	return ok && waitStatus.Signaled()
}

func transformResponse(response *http.Response) error {
	if response.Request.Method == http.MethodHead || response.StatusCode < 200 || response.StatusCode >= 300 || !strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "text/html") {
		return nil
	}
	if response.Body == nil {
		return nil
	}
	if encoding := response.Header.Get("Content-Encoding"); encoding != "" && !strings.EqualFold(encoding, "identity") {
		return fmt.Errorf("Hugo sent compressed HTML despite Accept-Encoding being disabled: %s", encoding)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read Hugo HTML response: %w", err)
	}
	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("close Hugo HTML response: %w", err)
	}
	transformed, changed, err := transformHTML(body)
	if err != nil {
		return fmt.Errorf("transform Hugo HTML response: %w", err)
	}
	if !changed {
		response.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}
	response.Body = io.NopCloser(bytes.NewReader(transformed))
	response.ContentLength = int64(len(transformed))
	response.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformed)))
	for _, header := range []string{"ETag", "Content-MD5", "Digest", "Content-Encoding"} {
		response.Header.Del(header)
	}
	return nil
}
