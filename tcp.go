package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"tailscale.com/tsnet"
)

func fwdTCP(lstConn net.Conn, ts *tsnet.Server, targetAddr string) error {
	defer lstConn.Close()

	// Enable TCP keepalive on listener connection
	if tcpConn, ok := lstConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsConn, err := ts.Dial(ctx, "tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("failed to dial tailscale node: %w", err)
	}
	defer tsConn.Close()

	// Enable TCP keepalive on Tailscale connection
	if tcpConn, ok := tsConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	var g errgroup.Group

	// client -> tailscale
	g.Go(func() error {
		defer func() {
			if tcpConn, ok := tsConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
		}()
		_, err := io.Copy(tsConn, lstConn)
		if err != nil {
			return fmt.Errorf("failed to copy data to tailscale node: %w", err)
		}
		return nil
	})

	// tailscale -> client
	g.Go(func() error {
		defer func() {
			if tcpConn, ok := lstConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
		}()
		_, err := io.Copy(lstConn, tsConn)
		if err != nil {
			return fmt.Errorf("failed to copy data from tailscale node: %w", err)
		}
		return nil
	})

	return g.Wait()
}
