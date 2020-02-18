go-gelf - GELF Library and Writer for Go
========================================

[GELF] (Graylog Extended Log Format) is an application-level logging
protocol that avoids many of the shortcomings of [syslog]. While it
can be run over any stream or datagram transport protocol, it has
special support ([chunking]) to allow long messages to be split over
multiple datagrams.

Versions
--------

In order to enable versionning of this package with Go, this project
is using GoPkg.in. The default branch of this project will be v1
for some time to prevent breaking clients. We encourage all project
to change their imports to the new GoPkg.in URIs as soon as possible.

To see up to date code, make sure to switch to the master branch.

v1.0.0
------

This implementation currently supports UDP and TCP as a transport
protocol. TLS is unsupported.

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

    go get gopkg.in/Graylog2/go-gelf.v1/gelf

    or

    go get github.com/Graylog2/go-gelf/gelf

This will get you version 1.0.0, with only UDP support and legacy API.
Newer versions are available through GoPkg.in:

    go get gopkg.in/Graylog2/go-gelf.v2/gelf

Usage
-----

The easiest way to integrate graylog logging into your go app is by
having your `main` function (or even `init`) call `log.SetOutput()`.
By using an `io.MultiWriter`, we can log to both stdout and graylog -
giving us both centralized and local logs.  (Redundancy is nice).

```golang
package main

import (
  "flag"
  "gopkg.in/Graylog2/go-gelf.v2/gelf"
  "io"
  "log"
  "os"
)

func main() {
  var graylogAddr string

  flag.StringVar(&graylogAddr, "graylog", "", "graylog server addr")
  flag.Parse()

  if graylogAddr != "" {
          // If using UDP
    gelfWriter, err := gelf.NewUDPWriter(graylogAddr)
          // If using TCP
          //gelfWriter, err := gelf.NewTCPWriter(graylogAddr)
    if err != nil {
      log.Fatalf("gelf.NewWriter: %s", err)
    }
    // log to both stderr and graylog2
    log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
    log.Printf("logging to stderr & graylog2@'%s'", graylogAddr)
  }

  // From here on out, any calls to log.Print* functions
  // will appear on stdout, and be sent over UDP or TCP to the
  // specified Graylog2 server.

  log.Printf("Hello gray World")

  // ...
}
```
The above program can be invoked as:

    go run test.go -graylog=localhost:12201

When using UDP messages may be dropped or re-ordered. However, Graylog
server availability will not impact application performance; there is
a small, fixed overhead per log call regardless of whether the target
server is reachable or not.

### WriteMessage Example

```golang
/*
This is a very basic example of sending custom messages to Graylog with the go-gelf package
GELF Spec https://docs.graylog.org/en/3.2/pages/gelf.html#gelf-via-udp
*/

package main

import (
	"fmt"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"os"
	"runtime"
	"time"
)

// GelfWriter is a pointer to a gelf.UPDWriter
var GelfWriter *gelf.UDPWriter

// GelfHostname is used across GELF messages to provide the source field
var GelfHostname string

func main() {
	fmt.Println("Starting tests")
	GelfHostname, _ = os.Hostname()
	createLogger()
	logMsg()
}

func createLogger() {
	GelfWriter, _ = gelf.NewUDPWriter("172.28.253.58:11589")

}

func wrapBuildMessage(s string, f string, l int32, ex map[string]interface{}) *gelf.Message {
	/*
		Level is a stanard syslog level
		Facility is deprecated
		Line is deprecated
		File is deprecated
	*/

	m := &gelf.Message{
		Version:  "1.1",
		Host:     GelfHostname,
		Short:    s,
		Full:     f,
		TimeUnix: float64(time.Now().Unix()),
		Level:    l,
		Extra:    ex,
	}

	return m
}

func LogMsg() {

	var level int32

	shortMsg := "Error"
	fullMsg := "Stack trace here"
	level = 3 // https://success.trendmicro.com/solution/TP000086250-What-are-Syslog-Facilities-and-Levels
	_, file, line, _ := runtime.Caller(1)
	customExtras := map[string]interface{}{"file": file, "line": line, "job_name": "testJob"}

	customMessage := wrapBuildMessage(shortMsg, fullMsg, level, customExtras)

	e := GelfWriter.WriteMessage(customMessage)

	if e != nil {
		fmt.Println("Received error when sending GELF message:", e.Error())
	} else {
		fmt.Println("Sent message")
	}

}
```

To Do
-----

- ~~WriteMessage example~~

License
-------

go-gelf is offered under the MIT license, see LICENSE for details.
