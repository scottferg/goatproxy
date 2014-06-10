GoatProxy
=========

A small HTTP proxy written in Go.

This tool is still very much pre-alpha, and won't work for anything more than light REST proxying.

## Usage

        $ go get
        $ go build
        $ ./goatproxy -host=my.app.com -listen_port=8888 -port=80 -with_ui
