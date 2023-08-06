package options

import (
	"flag"
	"os"
)

const (
	envAddress  = "ADDRESS"
	flagAddress = "a"
)

type Options struct {
	Host string
}

func Parse() Options {
	var opts = Options{}
	checkFlags(&opts)
	checkEnvironments(&opts)
	return opts
}

func checkEnvironments(opts *Options) {
	if envHost := os.Getenv(envAddress); envHost != "" {
		opts.Host = envHost
	}
}

func checkFlags(opts *Options) {
	flag.StringVar(&opts.Host, flagAddress, "localhost:8080", "server endpoint")
	flag.Parse()
}
