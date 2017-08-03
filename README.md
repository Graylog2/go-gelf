go-gelf - GELF Library and Writer for Go
========================================

[![Join the chat at https://gitter.im/Graylog2/go-gelf.svg](https://badges.gitter.im/Graylog2/go-gelf.svg)](https://gitter.im/Graylog2/go-gelf?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[GELF] (Graylog Extended Log Format) is an application-level logging
protocol that avoids many of the shortcomings of [syslog]. While it
can be run over any stream or datagram transport protocol, it has
special support ([chunking]) to allow long messages to be split over
multiple datagrams.

This implementation currently supports only UDP as a transport
protocol. TCP and TLS are unsupported.

The library provides an API that applications can use to log messages
directly to a Graylog server and an `io.Writer` that can be used to
redirect the standard library's log messages (`os.Stdout`) to a
Graylog server.

[GELF]: http://docs.graylog.org/en/2.2/pages/gelf.html
[syslog]: https://tools.ietf.org/html/rfc5424
[chunking]: http://docs.graylog.org/en/2.2/pages/gelf.html#chunked-gelf


Installing
----------

go-gelf is go get-able:

	go get github.com/Graylog2/go-gelf/gelf

Usage
-----

The easiest way to integrate graylog logging into your go app is by
having your `main` function (or even `init`) call `log.SetOutput()`.
By using an `io.MultiWriter`, we can log to both stdout and graylog -
giving us both centralized and local logs.  (Redundancy is nice).

	package main

	import (
		"flag"
		"github.com/Graylog2/go-gelf/gelf"
		"io"
		"log"
		"os"
	)

	func main() {
		var graylogAddr string

		flag.StringVar(&graylogAddr, "graylog", "", "graylog server addr")
		flag.Parse()

		if graylogAddr != "" {
			gelfWriter, err := gelf.NewWriter(graylogAddr)
			if err != nil {
				log.Fatalf("gelf.NewWriter: %s", err)
			}
			// log to both stderr and graylog2
			log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
			log.Printf("logging to stderr & graylog2@'%s'", graylogAddr)
		}

		// From here on out, any calls to log.Print* functions
		// will appear on stdout, and be sent over UDP to the
		// specified Graylog2 server.

		log.Printf("Hello gray World")

		// ...
	}

The above program can be invoked as:

	go run test.go -graylog=localhost:12201

When using UDP messages may be dropped or re-ordered. However, Graylog
server availability will not impact application performance; there is
a small, fixed overhead per log call regardless of whether the target
server is reachable or not.


To Do
-----

- WriteMessage example

License
-------

go-gelf is offered under the MIT license, see LICENSE for details.
