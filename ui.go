package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type WebUI struct {
	websocket *websocket.Conn
	from      string
	to        string
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Event struct {
	Event interface{} `json:"httpev"`
	Body  string      `json:"body"`
}

func openUrl(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}

	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
}

func (d *WebUI) Init() {
	go d.RunServer()
}

func readResponseBody(body []byte, contentType string) string {
	var objmap interface{}

	switch contentType {
	case "application/json":
		fmt.Println("JSON response")
		_ = json.Unmarshal(body, &objmap)

		res, _ := json.MarshalIndent(objmap, "", "    ")
		return string(res)
	case "text/html":
		return string(body)
	case "application/xml":
		_ = xml.Unmarshal(body, &objmap)

		res, _ := xml.MarshalIndent(objmap, "", "    ")
		return string(res)
	}

	return ""
}

func (d *WebUI) Update(op interface{}, body []byte) {
	if d.websocket != nil {
		e := Event{
			Event: op,
		}

		if v, ok := op.(*http.Response); ok {
			var contentType string
			if c, ok := v.Header["Content-Type"]; ok {
				contentType = strings.Split(c[0], ";")[0]
			} else {
				contentType = "text/plain"
			}

			e.Body = readResponseBody(body, contentType)
		} else if v, ok := op.(*http.Request); ok {
			var contentType string
			if c, ok := v.Header["Content-Type"]; ok {
				contentType = strings.Split(c[0], ";")[0]
			} else {
				contentType = "text/plain"
			}

			e.Body = readResponseBody(body, contentType)
		}

		websocket.JSON.Send(d.websocket, Message{
			Type:    "proxy-update",
			Payload: e,
		})
	}
}

func (d *WebUI) SockServer(ws *websocket.Conn) {
	var err error
	var clientMessage string

	d.websocket = ws

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			fmt.Println("Websocket could not be closed", err.Error())
		}
	}()
	client := ws.Request().RemoteAddr
	fmt.Println("Client connected:", client)

	websocket.JSON.Send(d.websocket, Message{
		Type:    "connect-success",
		Payload: fmt.Sprintf("Listening on port %s, forwarding to %s\n", d.from, d.to),
	})

	for {
		if err = websocket.Message.Receive(ws, &clientMessage); err != nil {
			return
		}
	}
}

func (d *WebUI) RunServer() {
	http.Handle("/js/", http.FileServer(http.Dir("./static/")))
	http.Handle("/font/", http.FileServer(http.Dir("./static/")))
	http.Handle("/css/", http.FileServer(http.Dir("./static/")))
	http.Handle("/socket", websocket.Handler(d.SockServer))
	http.Handle("/", http.FileServer(http.Dir("./static/html/")))

	openUrl("http://localhost:5000")

	err := http.ListenAndServe(fmt.Sprintf(":%d", 5000), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
