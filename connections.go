package main

import (
	"fmt"
	"net"
	"errors"
	"net/url"
	"crypto/tls"
)

func Connect(u string) (net.Conn, error) {
	target, _ := url.Parse(u)

	port := target.Port()

	if port == "" && target.Scheme == "http" {
		port = "80"
	} 
	if port == "" && target.Scheme == "https" {
		port = "443"
	}

	dialable := fmt.Sprintf("%s:%s", target.Hostname(), port)

	logger.Debug("Attempting to dial: ", dialable)

	if target.Scheme == "http" {
		return net.Dial("tcp", dialable)
	}

	if target.Scheme == "https" {
		return tls.Dial("tcp", dialable, &tls.Config{InsecureSkipVerify: true})
	}

	return nil, errors.New("Unable to create valid connection")
}