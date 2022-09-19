package main

import (
	"time"

	"github.com/miekg/dns"
)

type Server struct {
	listen   string
	rTimeout time.Duration
	wTimeout time.Duration
}

func (s *Server) Run() {
	h := NewHandler()

	th := dns.NewServeMux()
	th.HandleFunc(".", h.DoTCP)
	ts := &dns.Server{Addr: s.listen, Net: "tcp", Handler: th, ReadTimeout: s.rTimeout, WriteTimeout: s.wTimeout}
	go s.start(ts)

	uh := dns.NewServeMux()
	uh.HandleFunc(".", h.DoUDP)
	us := &dns.Server{Addr: s.listen, Net: "udp", Handler: uh, UDPSize: 65535, ReadTimeout: s.rTimeout, WriteTimeout: s.wTimeout}
	go s.start(us)
}

func (s *Server) start(ds *dns.Server) {
	logger.Info("Start %s listener on %s", ds.Net, s.listen)
	if err := ds.ListenAndServe(); err != nil {
		logger.Error("Start %s listener on %s failed:%s", ds.Net, s.listen, err.Error())
	}
}
