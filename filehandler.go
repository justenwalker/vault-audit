package main

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
)

type FileHandler struct {
	fd *lumberjack.Logger
}

func (f *FileHandler) Handle(line string) {
	_, err := f.fd.Write([]byte(line))
	if err != nil {
		log.Println("write error:", err)
	}
}

func (f *FileHandler) Close() error {
	return f.fd.Close()
}

func NewFileHandler(filename string) (*FileHandler, error) {
	log := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, //MB
		MaxBackups: 3,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	}
	if err := log.Rotate(); err != nil {
		return nil, err
	}
	return &FileHandler{
		fd: log,
	}, nil
}
