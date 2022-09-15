package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type ResolvError struct {
	qname, net  string
	nameservers []string
}

func (e ResolvError) Error() string {
	errmsg := fmt.Sprintf("%s resolv failed on %s (%s)", e.qname, strings.Join(e.nameservers, "; "), e.net)
	return errmsg
}

type RResp struct {
	msg        *dns.Msg
	nameserver string
	rtt        time.Duration
}

type Resolver struct {
	servers      []string
	domainServer *suffixTreeNode
	config       *ResolvConf
}

func NewResolver(c ResolvConf) *Resolver {
	r := &Resolver{
		domainServer: newSuffixTreeRoot(),
		config:       &c,
	}

	if c.ServerListFile != "" {
		r.ReadServerListFile(c.ServerListFile)
	}

	if c.ResolvFile != "" {
		clientConfig, err := dns.ClientConfigFromFile(c.ResolvFile)
		if err != nil {
			logger.Error(":%s is not a valid resolv.conf file\n", c.ResolvFile)
			logger.Error("%s", err)
			panic(err)
		}
		for _, server := range clientConfig.Servers {
			nameserver := net.JoinHostPort(server, clientConfig.Port)
			r.servers = append(r.servers, nameserver)
		}
	}

	return r
}

func (r *Resolver) parseServerListFile(buf *os.File) {
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "server") {
			continue
		}

		sli := strings.Split(line, "=")
		if len(sli) != 2 {
			continue
		}

		line = strings.TrimSpace(sli[1])

		tokens := strings.Split(line, "/")
		switch len(tokens) {
		case 3:
			domain := tokens[1]
			ip := tokens[2]

			if !isDomain(domain) || !isIP(ip) {
				continue
			}
			r.domainServer.sinsert(strings.Split(domain, "."), ip)
		case 1:
			srvPort := strings.Split(line, "#")
			if len(srvPort) > 2 {
				continue
			}

			ip := ""
			if ip = srvPort[0]; !isIP(ip) {
				continue
			}

			port := "53"
			if len(srvPort) == 2 {
				if _, err := strconv.Atoi(srvPort[1]); err != nil {
					continue
				}
				port = srvPort[1]
			}
			r.servers = append(r.servers, net.JoinHostPort(ip, port))
		}
	}
}

func (r *Resolver) ReadServerListFile(path string) {
	files := strings.Split(path, ";")
	for _, file := range files {
		buf, err := os.Open(file)
		if err != nil {
			panic("Can't open " + file)
		}
		defer buf.Close()
		r.parseServerListFile(buf)
	}
}

// Lookup will ask each nameserver in top-to-bottom fashion, starting a new request
// in every second, and return as early as possbile (have an answer).
// It returns an error if no request has succeeded.
func (r *Resolver) Lookup(net string, req *dns.Msg) (message *dns.Msg, err error) {
	c := &dns.Client{
		Net:          net,
		ReadTimeout:  r.Timeout(),
		WriteTimeout: r.Timeout(),
	}

	if net == "udp" && conf.ResolvConfig.SetEDNS0 {
		req = req.SetEdns0(65535, true)
	}

	qname := req.Question[0].Name

	res := make(chan *RResp, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()
		r, rtt, err := c.Exchange(req, nameserver)
		if err != nil {
			logger.Warn("%s socket error on %s", qname, nameserver)
			logger.Warn("error:%s", err.Error())
			return
		}
		// If SERVFAIL happen, should return immediately and try another upstream resolver.
		// However, other Error code like NXDOMAIN is an clear response stating
		// that it has been verified no such domain existas and ask other resolvers
		// would make no sense. See more about #20
		if r != nil && r.Rcode != dns.RcodeSuccess {
			logger.Warn("%s failed to get an valid answer on %s", qname, nameserver)
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		}
		re := &RResp{r, nameserver, rtt}
		select {
		case res <- re:
		default:
		}
	}

	ticker := time.NewTicker(time.Duration(conf.ResolvConfig.Interval) * time.Millisecond)
	defer ticker.Stop()
	// Start lookup on each nameserver top-down, in every second
	nameservers := r.Nameservers(qname)
	for _, nameserver := range nameservers {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case re := <-res:
			logger.Debug("%s resolv on %s rtt: %v", UnFqdn(qname), re.nameserver, re.rtt)
			return re.msg, nil
		case <-ticker.C:
			continue
		}
	}
	// wait for all the namservers to finish
	wg.Wait()
	select {
	case re := <-res:
		logger.Debug("%s resolv on %s rtt: %v", UnFqdn(qname), re.nameserver, re.rtt)
		return re.msg, nil
	default:
		return nil, ResolvError{qname, net, nameservers}
	}
}

// Namservers return the array of nameservers, with port number appended.
// '#' in the name is treated as port separator, as with dnsmasq.

func (r *Resolver) Nameservers(qname string) []string {
	queryKeys := strings.Split(qname, ".")
	queryKeys = queryKeys[:len(queryKeys)-1] // ignore last '.'

	var ns []string
	if v, found := r.domainServer.search(queryKeys); found {
		logger.Debug("%s be found in domain server list, upstream: %v", qname, v)
		server := v
		nameserver := net.JoinHostPort(server, "53")
		ns = append(ns, nameserver)
		// Ensure query the specific upstream nameserver in async Lookup() function.
		return ns
	}

	for _, nameserver := range r.servers {
		ns = append(ns, nameserver)
	}
	return ns
}

func (r *Resolver) Timeout() time.Duration {
	return time.Duration(r.config.Timeout) * time.Second
}
