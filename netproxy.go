package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	host        *string = flag.String("host", "", "target host or address")
	port        *string = flag.String("port", "0", "target port")
	listen_port *string = flag.String("listen_port", "0", "listen port")
	with_ui     *bool   = flag.Bool("with_ui", false, "launch web UI")

	ui *WebUI
)

func runUiServer() {
	ui = &WebUI{
		from: *listen_port,
		to:   fmt.Sprintf("%s:%s", *host, *port),
	}
	ui.Init()

	for {
		select {}
	}
}

func doForwardTraffic(in chan *http.Request, out chan *http.Response, ack chan interface{}) {
	host := fmt.Sprintf("%s:%s", *host, *port)
	client := &http.Client{}

	var resp *http.Response
	for {
		req := <-in
		req.RequestURI = ""

		req.ParseMultipartForm(0)

		var err error
		if req.Method == "POST" {
			resp, err = http.PostForm(fmt.Sprintf("http://%s%s", host, req.URL.RequestURI()), req.Form)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			var forwardReq *http.Request

			forwardReq, err = http.NewRequest(req.Method, fmt.Sprintf("http:%s", req.URL.String()), nil)
			forwardReq.URL.Host = host
			forwardReq.Header = req.Header

			resp, err = client.Do(forwardReq)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		ui.Update(req, nil)

		out <- resp
		ack <- req
	}
}

func doListenTraffic(in chan *http.Response, out chan *http.Request, ack chan interface{}) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		out <- r

		resp := <-in

		reader := io.TeeReader(resp.Body, w)
		body, _ := ioutil.ReadAll(reader)

		ui.Update(resp, body)

		ack <- resp
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", *listen_port), mux)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func main() {
	flag.Parse()

	if flag.NFlag() < 3 {
		fmt.Println("usage: netproxy -host target_host -port target_port -listen_port listen_port")
		flag.PrintDefaults()
		os.Exit(1)
	}

	closer := make(chan interface{})
	incomingFrom := make(chan *http.Request)
	outgoingFrom := make(chan *http.Response)

	go doListenTraffic(outgoingFrom, incomingFrom, closer)

	incomingTo := make(chan *http.Request)
	outgoingTo := make(chan *http.Response)

	go doForwardTraffic(incomingTo, outgoingTo, closer)

	fmt.Sprintf("Listening on port %s, forwarding to %s:%s\n", *listen_port, *host, *port)

	if *with_ui {
		go runUiServer()
	}

	for {
		select {
		case req := <-incomingFrom:
			incomingTo <- req
		case resp := <-outgoingTo:
			outgoingFrom <- resp
		case ack := <-closer:
			switch v := ack.(type) {
			case *http.Response:
				v.Body.Close()
			case *http.Request:
				v.Body.Close()
			}
		}
	}
}
