package buzzer

// This code has been copied and modified from
// https://github.com/gorilla/websocket/blob/master/examples/echo/server.go
// under license from The Gorilla WebScoket Authors.

// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var backend Server
var upgrader = websocket.Upgrader{} // use default options

func wsConnect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// StartWebServer creates a WebSocket-enabled, HTTP Server and listens at the
// specified endpoint. The client files should be passed to static.
func StartWebServer(server Server, endpoint, static string) {
	if static == "" {
		static = "./client"
	}

	backend = server

	log.SetFlags(0)
	// log.Print(static)
	http.HandleFunc("/ws", wsConnect)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(static))))
	http.Handle("/", http.RedirectHandler("/static/", http.StatusMovedPermanently))
	log.Fatal(http.ListenAndServe(endpoint, nil))
}
