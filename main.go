package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		log.Fatal("ERROR:", err)
	}
}

func run(ctx context.Context, args []string) (err error) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	var (
		addr string
		out  string
	)
	fs.StringVar(&addr, "addr", "unix:///tmp/audit.sock", "listen address for audit log events")
	fs.StringVar(&out, "out", "-", "output target for audit logs")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var handler func(line string)
	switch {
	case out == "-":
		handler = func(line string) {
			fmt.Print(line)
		}
	case strings.HasPrefix(out, "file://"):
		var (
			u  *url.URL
			fd *FileHandler
		)
		u, err = url.Parse(out)
		if err != nil {
			return err
		}
		fd, err = NewFileHandler(u.Path)
		if err != nil {
			return err
		}
		handler = fd.Handle
		defer func() {
			if cerr := fd.Close(); cerr != nil {
				if err == nil {
					err = cerr
				} else {
					err = fmt.Errorf("shutdown errors: %w, %w", err, cerr)
				}
			}
		}()
	}

	log.Println("Listen =>", addr)
	s, err := NewServer(addr, handler)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	go s.Serve()
	<-ctx.Done()
	return s.Stop(context.Background())
}
