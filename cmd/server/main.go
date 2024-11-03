package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Endpoints map[string]Endpoint
}

type Endpoint struct {
	Path     string
	Status   int
	Response Response
}

type Response struct {
	Text string
	Json string
}

func makeHandler(status int, resp *Response) func(http.ResponseWriter, *http.Request) {
	h := func(w http.ResponseWriter, req *http.Request) {
		accept := "text/plain"
		accepts := req.Header["Accept"]
		if len(accepts) > 0 {
			accept = accepts[0]
		}

		headers := w.Header()
		if accept == "application/json" {
			headers.Set("Accept", "application/json")
		}
		w.WriteHeader(status)

		if accept == "application/json" {
			fmt.Fprintf(w, "%s", resp.Json)
		} else {
			fmt.Fprintf(w, "%s", resp.Text)
		}
	}
	return h
}

func main() {
	f := "example.toml"
	var config Config
	_, err := toml.DecodeFile(f, &config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, v := range config.Endpoints {
		http.HandleFunc(fmt.Sprintf("/%s", v.Path), makeHandler(v.Status, &v.Response))
	}

	http.ListenAndServe(":8080", nil)
}
