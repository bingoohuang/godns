package main

import (
	"strconv"
)

var conf Conf

var LogLevelMap = map[string]int{
	"DEBUG":  LevelDebug,
	"INFO":   LevelInfo,
	"NOTICE": LevelNotice,
	"WARN":   LevelWarn,
	"ERROR":  LevelError,
}

type Conf struct {
	Version      string
	Debug        bool
	Server       DNSServerConf `toml:"server"`
	ResolvConfig ResolvConf    `toml:"resolv"`
	Redis        RedisConf     `toml:"redis"`
	Memcache     MemcacheConf  `toml:"memcache"`
	Log          LogConf       `toml:"log"`
	Cache        CacheConf     `toml:"cache"`
	Hosts        HostsConf     `toml:"hosts"`
}

type ResolvConf struct {
	Timeout        int
	Interval       int
	SetEDNS0       bool
	ServerListFile string `toml:"server-list-file"`
	ResolvFile     string `toml:"resolv-file"`
}

type DNSServerConf struct {
	Host string
	Port int
}

type RedisConf struct {
	Host     string
	Port     int
	DB       int
	Password string
}

type MemcacheConf struct {
	Servers []string
}

func (s RedisConf) Addr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

type LogConf struct {
	Stdout bool
	File   string
	Level  string
}

func (ls LogConf) LogLevel() int {
	l, ok := LogLevelMap[ls.Level]
	if !ok {
		panic("Config error: invalid log level: " + ls.Level)
	}
	return l
}

type CacheConf struct {
	Backend  string
	Expire   int
	MaxCount int `toml:"max-count"`
}

type HostsConf struct {
	Enable          bool
	HostsFile       string `toml:"host-file"`
	RedisEnable     bool   `toml:"redis-enable"`
	RedisKey        string `toml:"redis-key"`
	TTL             uint32 `toml:"ttl"`
	RefreshInterval uint32 `toml:"refresh-interval"`
}
