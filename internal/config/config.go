package config

import (
	"flag"
	//"fmt"
	"regexp"

	//"github.com/pterm/pterm"
	//"github.com/pterm/pterm/putils"
)

type Config struct {
	Addr           *string
	Port           *int
	DnsAddr        *string
	DnsPort        *int
	EnableDoh      *bool
	Debug          *bool
	NoBanner       *bool
	SystemProxy    *bool
	Timeout        *int
	AllowedPattern []*regexp.Regexp
	WindowSize     *int
	Version        *bool
}

func New() {
	config := &Config{}
	config.Addr = flag.String("addr", "127.0.0.1", "listen address")
	config.Port = flag.Int("port", 8080, "port")
	config.DnsAddr = flag.String("dns-addr", "8.8.8.8", "dns address")
	config.DnsPort = flag.Int("dns-port", 53, "port number for dns")
	config.EnableDoh = flag.Bool("enable-doh", false, "enable 'dns-over-https'")
	config.Debug = flag.Bool("debug", false, "enable debug output")
	config.NoBanner = flag.Bool("no-banner", false, "disable banner")
	config.SystemProxy = flag.Bool("system-proxy", true, "enable system-wide proxy")
	config.Timeout = flag.Int("timeout", 0, "timeout in milliseconds; no timeout when not given")
	config.Version = flag.Bool("v", false, "print spoof-dpi's version; this may contain some other relevant information")
	flag.Parse()
}