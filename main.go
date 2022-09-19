package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/bingoohuang/gg/pkg/v"

	"github.com/BurntSushi/toml"
)

var logger *GoDNSLogger

func main() {
	configFile := flag.String("c", "./etc/godns.toml", "Look for godns toml-formatting config file in this directory")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	if _, err := toml.DecodeFile(*configFile, &conf); err != nil {
		log.Fatalf("%s is not a valid toml config file, error: %+v", *configFile, err)
	}

	if *verbose {
		conf.Log.Stdout = true
		conf.Log.Level = "DEBUG"
	}

	logger = newLogger()
	if *verbose {
		logger.Info("godns version: %s", v.Version())
	}

	server := &Server{
		listen:   conf.Server.Listen,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}

	server.Run()

	logger.Info("godns %s start")

	if conf.Debug {
		go profileCPU()
		go profileMEM()
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	<-sig
	logger.Info("signal received, stopping")
}

func profileCPU() {
	f, err := os.Create("godns.cprof")
	if err != nil {
		logger.Error("%s", err)
		return
	}

	pprof.StartCPUProfile(f)
	time.AfterFunc(6*time.Minute, func() {
		pprof.StopCPUProfile()
		f.Close()
	})
}

func profileMEM() {
	f, err := os.Create("godns.mprof")
	if err != nil {
		logger.Error("%s", err)
		return
	}

	time.AfterFunc(5*time.Minute, func() {
		pprof.WriteHeapProfile(f)
		f.Close()
	})
}

func newLogger() *GoDNSLogger {
	l := NewLogger()

	if conf.Log.Stdout {
		l.SetLogger("console", nil)
	}

	if conf.Log.File != "" {
		config := map[string]interface{}{"file": conf.Log.File}
		l.SetLogger("file", config)
	}

	l.SetLevel(conf.Log.LogLevel())
	return l
}
