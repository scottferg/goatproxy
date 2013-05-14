package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func colored(text, color string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", color, text)
}

func yellow(text string) string {
	return colored(text, "33")
}

func yellowBold(text string) string {
	return colored(text, "1;33")
}

func green(text string) string {
	return colored(text, "32")
}

func blue(text string) string {
	return colored(text, "34")
}

func blueBold(text string) string {
	return colored(text, "1;34")
}

func red(text string) string {
	return colored(text, "31")
}

func redBold(text string) string {
	return colored(text, "1;31")
}

func cyan(text string) string {
	return colored(text, "36")
}

func cyanBold(text string) string {
	return colored(text, "1;36")
}

func renderBody(r *http.Response, buf *bytes.Buffer) {
	buf.Write([]byte("\r\n"))

	var objmap interface{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	contentType := strings.Split(r.Header["Content-Type"][0], ";")
	switch contentType[0] {
	case "application/json":
		_ = json.Unmarshal(body, &objmap)

		res, _ := json.MarshalIndent(objmap, "", "    ")
		buf.Write(res)
	case "text/html":
		buf.Write(body)
	case "application/xml":
		_ = xml.Unmarshal(body, &objmap)

		res, _ := xml.MarshalIndent(objmap, "", "    ")
		buf.Write(res)
	}

	buf.Write([]byte("\r\n"))
}

func formattedRequest(r *http.Request) string {
	buf := new(bytes.Buffer)

	buf.Write([]byte(fmt.Sprintf("%s %s %s\n", green(r.Method), r.URL.String(), r.Proto)))
	buf.Write([]byte(fmt.Sprintf("%s: %s\n", green("Host"), r.Host)))

	if r.Header != nil {
		for k, v := range r.Header {
			buf.Write([]byte(fmt.Sprintf("%s: %s\n", green(k), v[0])))
		}
	}

	return string(buf.Bytes())
}

func formattedResponse(r *http.Response) string {
	buf := new(bytes.Buffer)

	switch {
	case r.StatusCode >= 500 && r.StatusCode < 600:
		buf.Write([]byte(fmt.Sprintf("%s %s\n", r.Proto, redBold(r.Status))))
	default:
		buf.Write([]byte(fmt.Sprintf("%s %s\n", r.Proto, cyanBold(r.Status))))
	}

	if r.Header != nil {
		for k, v := range r.Header {
			buf.Write([]byte(fmt.Sprintf("%s: %s\n", green(k), v[0])))
		}
	}

	if r.Header["Content-Type"] != nil && !*with_ui {
		renderBody(r, buf)
	}

	return string(buf.Bytes())
}
