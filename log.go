package main

import (
	"fmt"
	"log"
	"os"
)

const LogOutputBuffer = 1024

const (
	LevelDebug = iota
	LevelInfo
	LevelNotice
	LevelWarn
	LevelError
)

type logMsg struct {
	Level int
	Msg   string
}

type LoggerHandler interface {
	Setup(config map[string]interface{}) error
	Write(msg *logMsg)
}

type GoDNSLogger struct {
	level   int
	msgChan chan *logMsg
	outputs map[string]LoggerHandler
}

func NewLogger() *GoDNSLogger {
	l := &GoDNSLogger{
		msgChan: make(chan *logMsg, LogOutputBuffer),
		outputs: make(map[string]LoggerHandler),
	}
	go l.Run()
	return l
}

func (l *GoDNSLogger) SetLogger(handlerType string, config map[string]interface{}) {
	var handler LoggerHandler
	switch handlerType {
	case "console":
		handler = NewConsoleHandler()
	case "file":
		handler = NewFileHandler()
	default:
		panic("Unknown log handler.")
	}

	handler.Setup(config)
	l.outputs[handlerType] = handler
}

func (l *GoDNSLogger) SetLevel(level int) {
	l.level = level
}

func (l *GoDNSLogger) Run() {
	for {
		select {
		case m := <-l.msgChan:
			for _, handler := range l.outputs {
				handler.Write(m)
			}
		}
	}
}

func (l *GoDNSLogger) writeMsg(msg string, level int) {
	if l.level > level {
		return
	}

	lm := &logMsg{
		Level: level,
		Msg:   msg,
	}

	l.msgChan <- lm
}

func (l *GoDNSLogger) Debug(format string, v ...interface{}) {
	m := fmt.Sprintf("[DEBUG] "+format, v...)
	l.writeMsg(m, LevelDebug)
}

func (l *GoDNSLogger) Info(format string, v ...interface{}) {
	m := fmt.Sprintf("[INFO] "+format, v...)
	l.writeMsg(m, LevelInfo)
}

func (l *GoDNSLogger) Notice(format string, v ...interface{}) {
	m := fmt.Sprintf("[NOTICE] "+format, v...)
	l.writeMsg(m, LevelNotice)
}

func (l *GoDNSLogger) Warn(format string, v ...interface{}) {
	m := fmt.Sprintf("[WARN] "+format, v...)
	l.writeMsg(m, LevelWarn)
}

func (l *GoDNSLogger) Error(format string, v ...interface{}) {
	m := fmt.Sprintf("[ERROR] "+format, v...)
	l.writeMsg(m, LevelError)
}

type ConsoleHandler struct {
	level  int
	logger *log.Logger
}

func NewConsoleHandler() LoggerHandler {
	return new(ConsoleHandler)
}

func (h *ConsoleHandler) Setup(config map[string]interface{}) error {
	if _level, ok := config["level"]; ok {
		level := _level.(int)
		h.level = level
	}
	h.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return nil
}

func (h *ConsoleHandler) Write(lm *logMsg) {
	if h.level <= lm.Level {
		h.logger.Println(lm.Msg)
	}
}

type FileHandler struct {
	level  int
	file   string
	logger *log.Logger
}

func NewFileHandler() LoggerHandler {
	return new(FileHandler)
}

func (h *FileHandler) Setup(config map[string]interface{}) error {
	if level, ok := config["level"]; ok {
		h.level = level.(int)
	}

	if file, ok := config["file"]; ok {
		h.file = file.(string)
		output, err := os.OpenFile(h.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}

		h.logger = log.New(output, "", log.Ldate|log.Ltime)
	}

	return nil
}

func (h *FileHandler) Write(lm *logMsg) {
	if h.logger == nil {
		return
	}

	if h.level <= lm.Level {
		h.logger.Println(lm.Msg)
	}
}
