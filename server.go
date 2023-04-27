package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"sync"
	"time"
)

type Server struct {
	addr      net.Addr
	quitCh    chan struct{}
	handlerCh chan struct{}
	handler   func(line string)
	listener  net.Listener
	stopOnce  sync.Once
	stopErr   error
}

type HandlerFunc func(line string)

func NewServer(address string, handler HandlerFunc) (*Server, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("could not parse address '%s': %w", address, err)
	}
	var addr net.Addr
	switch u.Scheme {
	case "unix":
		addr, err = net.ResolveUnixAddr(u.Scheme, u.Path)
	case "tcp", "tcp4", "tcp6":
		addr, err = net.ResolveTCPAddr(u.Scheme, u.Host)
	case "udp", "upd4", "udp6":
		addr, err = net.ResolveUDPAddr(u.Scheme, u.Host)
	default:
		return nil, fmt.Errorf("unsupported network type: %s", u.Scheme)
	}
	if err != nil {
		return nil, fmt.Errorf("could not resolve address '%s': %w", address, err)
	}
	if addr.Network() == "unix" {
		if err := removeFile(addr.String()); err != nil {
			return nil, fmt.Errorf("error removing UDS '%s': %w", addr.String(), err)
		}
	}
	ln, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		return nil, fmt.Errorf("listen error on '%s': %w", address, err)
	}
	return &Server{
		addr:      addr,
		quitCh:    make(chan struct{}),
		handlerCh: make(chan struct{}, 10),
		handler:   handler,
		listener:  ln,
	}, nil
}

func removeFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return os.Remove(filename)
}

func (s *Server) Serve() {
	defer func() {
		if err := s.listener.Close(); err != nil {
			log.Printf("error closing listener [%s]: %v", s.addr, err)
		}
	}()
	for {
		select {
		case <-s.quitCh:
			// close listener
			return
		default:
			// accept next connection
		}
		conn, err := s.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		go s.handle(conn)
	}

}

func (s *Server) cleanup() error {
	if s.addr.Network() != "unix" {
		return nil
	}
	return removeFile(s.addr.String())
}

func (s *Server) handle(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	select {
	case s.handlerCh <- struct{}{}:
		// acquired
	case <-s.quitCh:
		// shutting down
		return
	case <-time.After(1 * time.Second):
		return // took too long to acquire
	}
	// after we're done, release the semaphore
	defer func() {
		<-s.handlerCh
	}()
	bufr := bufio.NewReader(conn)
	for {
		select {
		case <-s.quitCh:
			return // shut down
		default:
			// continue and read line
		}
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		line, err := bufr.ReadString('\n')
		if line != "" {
			s.handler(line)
		}
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue // try again
		}
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			log.Printf("read error: %v", err)
			return
		}
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.stopOnce.Do(func() {
		close(s.quitCh)
		for i := 0; i < cap(s.handlerCh); i++ {
			select {
			case s.handlerCh <- struct{}{}:
				// handled
			case <-ctx.Done():
				_ = s.cleanup()
				s.stopErr = ctx.Err()
				return
			}
			s.stopErr = s.cleanup()
		}
	})
	return s.stopErr
}

func (s *Server) Close() error {
	return s.Stop(context.Background())
}
