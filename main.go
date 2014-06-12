package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	host    *string = flag.String("host", "", "target host or address")
	port    *string = flag.String("port", ":8080", "target port")
	latency *int    = flag.Int("latency", 0, "in milliseconds")
)

type errorHandler func(w http.ResponseWriter, r *http.Request) error

func (e errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := e(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func prettyPrintBody(body *bytes.Buffer) ([]byte, error) {
	buf := &bytes.Buffer{}

	if len(body.Bytes()) > 0 {
		err := json.Indent(buf, body.Bytes(), "", "    ")
		if err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	return nil, nil
}

func prettyPrintRequest(r *http.Request, content string, body *bytes.Buffer) error {
	fmt.Printf("%s %s %s\n", GreenBold(r.Method), r.URL.String(), Blue(r.Proto))

	for header, values := range r.Header {
		fmt.Printf("%s: %s\n", YellowBold(header), strings.Join(values, ", "))
	}

	var output []byte
	var err error

	switch content {
	case "application/json":
		output, err = prettyPrintBody(body)
		if err != nil {
			return err
		}
	default:
		output = body.Bytes()
	}

	if len(output) > 0 {
		fmt.Println(string(output))
	}

	fmt.Printf("\n\n")
	return nil
}

func prettyPrintResponse(r *http.Response, content string, body *bytes.Buffer) error {
	fmt.Printf("%s %s\n", GreenBold(r.Proto), Blue(r.Status))

	for header, values := range r.Header {
		fmt.Printf("%s: %s\n", YellowBold(header), strings.Join(values, ", "))
	}

	var output []byte
	var err error

	switch content {
	case "application/json":
		output, err = prettyPrintBody(body)
		if err != nil {
			return err
		}
	default:
		output = body.Bytes()
	}

	fmt.Printf("%s\n\n\n", string(output))
	return nil
}

func proxyHandler(w http.ResponseWriter, r *http.Request) error {
	buf := new(bytes.Buffer)
	reader := io.TeeReader(r.Body, buf)
	defer r.Body.Close()

	// Repair URI if necessary
	uri := r.URL
	if uri.Scheme == "" {
		uri.Scheme = "http"
	}

	if uri.Host == "" {
		uri.Host = *host
	}

	req, err := http.NewRequest(r.Method, fmt.Sprintf("%s", uri.String()), reader)
	if err != nil {
		return err
	}
	req.Header = r.Header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Simulate latency
	time.Sleep(time.Duration(*latency) * time.Millisecond)

	err = prettyPrintRequest(r, resp.Header["Content-Type"][0], buf)
	if err != nil {
		return err
	}

	buf = bytes.NewBuffer(nil)
	reader = io.TeeReader(resp.Body, buf)
	defer resp.Body.Close()

	for header, values := range resp.Header {
		w.Header().Set(header, strings.Join(values, ", "))
	}

	_, err = bufio.NewWriter(w).ReadFrom(reader)

	switch {
	case len(resp.Header["Content-Encoding"]) > 0 && resp.Header["Content-Encoding"][0] == "gzip":
		uncom, err := gzip.NewReader(buf)
		if err != nil {
			return err
		}

		buf = bytes.NewBuffer(nil)
		_, err = buf.ReadFrom(uncom)
		if err != nil {
			return err
		}
	}

	err = prettyPrintResponse(resp, resp.Header["Content-Type"][0], buf)
	if err != nil {
		return err
	}

	return err
}

func init() {
	flag.Parse()

	if flag.NFlag() < 1 {
		fmt.Println("usage: netproxy -host target_host")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	http.Handle("/", errorHandler(proxyHandler))

	log.Fatal(http.ListenAndServe(*port, nil))
}
