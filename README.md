# GODNS

A simple and fast dns cache server written by go. fork from [kenshinx/godns](https://github.com/kenshinx/godns).

Similar to [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html), but supports some difference features:

* Keep hosts records in redis and the local file /etc/hosts
* Auto-Reloads when hosts configuration is changed. (Yes, dnsmasq needs to be reloaded)

## Installation & Running

1. Install

    $ go get github.com/kenshinx/godns

2. Build

    $ cd $GOPATH/src/github.com/kenshinx/godns
    $ go build -o godns

3. Running

    $ sudo ./godns -c ./etc/godns.conf

4. Test

    $ dig www.github.com @127.0.0.1
    ```sh
    $ dig @101.4.1.17 wiki.bjca

    ; <<>> DiG 9.10.6 <<>> @101.4.1.17 wiki.bjca
    ; (1 server found)
    ;; global options: +cmd
    ;; Got answer:
    ;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 1578
    ;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
    ;; WARNING: recursion requested but not available
    
    ;; QUESTION SECTION:
    ;wiki.bjca.			IN	A
    
    ;; ANSWER SECTION:
    wiki.bjca.		600	IN	A	192.168.131.51
    
    ;; Query time: 4 msec
    ;; SERVER: 101.4.1.17#53(101.4.1.17)
    ;; WHEN: Mon Jul 25 16:36:20 CST 2022
    ;; MSG SIZE  rcvd: 52
    ```

## Use godns

    $ sudo vi /etc/resolv.conf
    nameserver #the ip of godns running

## Configuration

All the configuration in `godns.conf` is a TOML format config file.
More about Toml :[https://github.com/mojombo/toml](https://github.com/mojombo/toml)

### resolv.conf

Upstream server can be configured by changing file from somewhere other than "/etc/resolv.conf"

```toml
[resolv]
resolv-file = "/etc/resolv.conf"
```

If multiple `namerservers` are set in resolv.conf, the upsteam server will try in a top to bottom order

### server-list-file

Domain-specific nameservers configuration, formatting keep compatible with Dnsmasq.
>server=/google.com/8.8.8.8

More cases please refererence [dnsmasq-china-list](https://github.com/felixonmars/dnsmasq-china-list)

### cache

Only the local memory storage backend is currently implemented.  The redis backend is in the todo list

```toml
[cache]
backend = "memory"
expire = 600  # default expire time 10 minutes
maxcount = 100000
```

### hosts

Force resolve domain to assigned ip, support two types hosts configuration:

* locale hosts file
* remote redis hosts

__hosts file__

can be assigned at godns.conf,default : `/etc/hosts`

```toml
[hosts]
host-file = "/etc/hosts"
```

Hosts file format is described in [linux man pages](http://man7.org/linux/man-pages/man5/hosts.5.html).
More than that , `*.` wildcard is supported additional.

__redis hosts__

This is a special requirment in our system. Must maintain a global hosts configuration,
and support update the host records from other remote server.
Therefore, while "redis-hosts" be enabled, will query the redis db when each dns request is reached.

The hosts record is organized with redis hash map. and the key of the map is configured.

```toml
[hosts]
redis-key = "godns:hosts"
```

_Insert hosts records into redis_

```sh
redis > hset godns:hosts www.test.com 1.1.1.1
```

Compared with file-backend records, redis-backend hosts support multiple A entries.

```sh
redis > hset godns:hosts www.test.com 1.1.1.1,2.2.2.2
```

## Benchmark

__Debug close__

```sh
$ go test -bench=.

testing: warning: no tests to run
PASS
BenchmarkDig-8     50000             57945 ns/op
ok      _/usr/home/keqiang/godns        3.259s
```

The result : 15342 queries/per second

The test environment:

CentOS release 6.4

* CPU:
Intel Xeon 2.40GHZ
4 cores

* MEM:
46G

## Web console

Joke: A web console for godns

[https://github.com/kenshinx/joke](https://github.com/kenshinx/joke)

screenshot

![joke](https://raw.github.com/kenshinx/joke/master/screenshot/joke.png)

## Deployment

Deployment in productive supervisord highly recommended.

```sh

[program:godns]
command=/usr/local/bin/godns -c /etc/godns.conf
autostart=true
autorestart=true
user=root
stdout_logfile_maxbytes = 50MB
stdoiut_logfile_backups = 20
stdout_logfile = /var/log/godns.log

```

## TODO

* The redis cache backend
* Update ttl

## LICENSE

godns is under the MIT license. See the LICENSE file for details.
