package main

import (
	"net"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

type Server struct {
	host     string
	port     int
	rTimeout time.Duration
	wTimeout time.Duration
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.host, strconv.Itoa(s.port))
}

func (s *Server) Run() {
	h := NewHandler()

	th := dns.NewServeMux()
	th.HandleFunc(".", h.DoTCP)
	ts := &dns.Server{Addr: s.Addr(), Net: "tcp", Handler: th, ReadTimeout: s.rTimeout, WriteTimeout: s.wTimeout}
	go s.start(ts)

	uh := dns.NewServeMux()
	uh.HandleFunc(".", h.DoUDP)
	us := &dns.Server{Addr: s.Addr(), Net: "udp", Handler: uh, UDPSize: 65535, ReadTimeout: s.rTimeout, WriteTimeout: s.wTimeout}
	go s.start(us)
}

func (s *Server) start(ds *dns.Server) {
	logger.Info("Start %s listener on %s", ds.Net, s.Addr())
	if err := ds.ListenAndServe(); err != nil {
		logger.Error("Start %s listener on %s failed:%s", ds.Net, s.Addr(), err.Error())
	}
}
