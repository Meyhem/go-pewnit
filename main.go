package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
)

func Die(print ...interface{}) {
	for _, p := range print {
		fmt.Println(p)
	}

	os.Exit(1)
}

func AssertValidUrl(u string) {
	p, err := url.Parse(u)

	if err != nil {
		Die("Invalid URL provided", err)
	}

	if p.Scheme == "" {
		Die("Url is not fully qualified, provide scheme")
	}

	if p.Host == "" {
		Die("Url is not fully qualified, provide host name")
	}
}

var opts struct {
	LogLevel string `short:"l" long:"loglevel" description:"Sets log level" choice:"DEBUG" choice:"INFO" choice:"NOTICE" choice:"WARNING" choice:"ERROR" choice:"CRITICAL"`

	Concurrency uint `short:"c" long:"concurrency" default:"200" description:"Number of concurrent attacks"`

	AttackType string `short:"a" long:"attack" required:"true" description:"Specifies type of attack to execute" choice:"connectionflood" choice:"slowloris" choice:"httpflood"`

	Method string `short:"m" long:"method" default:"GET" description:"HTTP method to send to target server" choice:"GET" choice:"POST" choice:"PUT" choice:"DELETE" choice:"PATCH" choice:"HEAD" choice:"TRACE" choice:"CONNECT"`

	Body string `short:"b" long:"body" description:"Custom body to provide with request. Only applicable on valid method combinations (not GET, HEAD, TRACE). Must be correctly encoded by user and user must provide corresponding Content-Type header. Content-Length is provided automatically."`

	Header []string `long:"header" description:"Additional header to send. Option can be provided multiple times"`

	Porcelaine bool `long:"porcelaine" description:"Script friendly output"`

	PositionalArgs struct {
		URL string `positional-arg-name:"URL" description:"Url to attack"`
	} `positional-args:"yes" required:"true"`
}

var logger = logging.MustGetLogger("pewnit")

func main() {
	rand.Seed(time.Now().UnixNano())
	args := os.Args[1:]
	_, err := flags.ParseArgs(&opts, args)

	if err != nil {
		Die()
	}

	const format = "%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level} %{id:03x}%{color:reset} %{message}"
	const porcelaineFormat = "%{time:15:04:05.000} %{shortfunc} -> %{level} %{id:03x} %{message}"
	var logFormatter logging.Formatter
	if opts.Porcelaine {
		logFormatter, _ = logging.NewStringFormatter(porcelaineFormat)
	} else {
		logFormatter, _ = logging.NewStringFormatter(format)
	}

	logLevel := logging.WARNING

	switch opts.LogLevel {
	case "DEBUG":
		logLevel = logging.DEBUG
	case "INFO":
		logLevel = logging.INFO
	case "NOTICE":
		logLevel = logging.NOTICE
	case "WARNING":
		logLevel = logging.WARNING
	case "ERROR":
		logLevel = logging.ERROR
	case "CRITICAL":
		logLevel = logging.CRITICAL
	default:
		logLevel = logging.WARNING
	}

	logging.SetLevel(logLevel, "pewnit")
	logging.SetFormatter(logFormatter)

	if opts.PositionalArgs.URL == "" {
		fmt.Println("Positional argument <URL> not provided")
		Die()
	}

	if opts.Body != "" && (opts.Method == "GET" || opts.Method == "HEAD" || opts.Method == "TRACE") {
		logger.Warning("Using body data with GET, HEAD, TRACE methods is not valid HTTP. THis might diminish the attack.")
	}

	AssertValidUrl(opts.PositionalArgs.URL)

	engine := NewEngine(opts.PositionalArgs.URL, opts.Concurrency, opts.AttackType, opts.Method, opts.Body, opts.Header)
	engine.Attack()
}
