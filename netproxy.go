package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	host        *string = flag.String("host", "", "target host or address")
	port        *string = flag.String("port", "0", "target port")
	listen_port *string = flag.String("listen_port", "0", "listen port")
	with_ui     *bool   = flag.Bool("with_ui", false, "launch web UI")

	lastRequest *http.Request

	methods = []string{
		"GET",
		"PUT",
		"POST",
		"DELETE",
		"HEAD",
		"PATCH",
		"OPTIONS",
	}
)

type ProxyConnection struct {
	from, to net.Conn
	logger   chan interface{}
	ack      chan bool
}

func die(format string, v ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf(format+"\n", v...))
	os.Exit(1)
}

func renderPackets(buf *bytes.Buffer) interface{} {
	reader := bufio.NewReader(buf)

	for _, v := range methods {
		if strings.Index(string(buf.Bytes())[:10], v) > -1 {
			req, _ := http.ReadRequest(reader)
			lastRequest = req
			fmt.Println(formattedRequest(req))
			return req
		}
	}

	resp, _ := http.ReadResponse(reader, lastRequest)
	fmt.Println(formattedResponse(resp))
	return resp
}

func proxy(c *ProxyConnection) {
	/*
		from := c.from.LocalAddr().String()
		to := c.to.LocalAddr().String()
	*/

	b := make([]byte, 10246)
	buf := new(bytes.Buffer)

	var totalPackets int
	var totalBytes int

	for {
		numBytes, err := c.from.Read(b)
		if err != nil {
			break
		}

		if numBytes > 0 {
			buf.Write(b[:numBytes])
			c.to.Write(b[:numBytes])
			totalPackets++
		}

		totalBytes += numBytes
	}

	c.logger <- renderPackets(buf)
	// c.logger <- yellowBold(fmt.Sprintf("%d bytes sent from %s to %s\r\n", totalBytes, from, to))

	c.from.Close()
	c.to.Close()
	c.ack <- true
}

func runUiServer(logger chan interface{}, target string) {
	w := WebUI{
		from: *listen_port,
		to:   target,
	}
	w.Init()

	for {
		b := <-logger
		if b == nil {
			return
		}

		w.Update(b)
	}
}

func handleConnection(local net.Conn, target string, logger chan interface{}) {
	remote, err := net.Dial("tcp", target)
	if err != nil {
		fmt.Printf("Unable to connect to %s, %v\n", target, err)
	}

	// started := time.Now()

	reqack := make(chan bool)
	respack := make(chan bool)

	go proxy(&ProxyConnection{
		from:   remote,
		to:     local,
		logger: logger,
		ack:    respack,
	})

	go proxy(&ProxyConnection{
		from:   local,
		to:     remote,
		logger: logger,
		ack:    reqack,
	})

	<-reqack
	<-respack

	// finished := time.Now()
	// duration := finished.Sub(started)

	// logger <- fmt.Sprintf("Duration %s\n", duration.String())
}

func main() {
	flag.Parse()

	if flag.NFlag() < 3 {
		fmt.Println("usage: netproxy -host target_host -port target_port -listen_port listen_port")
		flag.PrintDefaults()
		os.Exit(1)
	}

	target := net.JoinHostPort(*host, *port)
	fmt.Printf("Listening on port %s, forwarding to %s\n", *listen_port, target)

	logger := make(chan interface{})

	if *with_ui {
		go runUiServer(logger, target)
	}

	ln, err := net.Listen("tcp", ":"+*listen_port)
	if err != nil {
		fmt.Printf("Unable to start listener, %v\n", err)
		os.Exit(1)
	}

	for {
		if conn, err := ln.Accept(); err == nil {
			go handleConnection(conn, target, logger)
		} else {
			fmt.Printf("Accept failed, %v\n", err)
		}
	}
}
