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
	Verbose []bool `short:"v" long:"verbose" description:"Log verbosity"`

	Concurrency uint `short:"c" long:"concurrency" default:"5" description:"Number of concurrent attacks"`

	AttackType string `short:"a" long:"attack" required:"true" description:"Specifies type of attack to execute" choice:"connectionflood" choice:"slowloris" choice:"httpflood"`

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

	logging.SetLevel(logging.ERROR, "pewnit")
	logging.SetFormatter(logFormatter)

	if opts.PositionalArgs.URL == "" {
		fmt.Println("Positional argument <URL> not provided")
		Die()
	}

	AssertValidUrl(opts.PositionalArgs.URL)

	engine := NewEngine(opts.PositionalArgs.URL, opts.Concurrency, opts.AttackType)
	engine.Attack()
}
