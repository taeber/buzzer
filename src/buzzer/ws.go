package buzzer

// This code has been copied and modified from
// https://github.com/gorilla/websocket/blob/master/examples/echo/server.go
// under license from The Gorilla WebScoket Authors.

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	username string
	socket   *websocket.Conn
	lock     sync.Mutex
}

var backend Server
var upgrader = websocket.Upgrader{} // use default options

func accept(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	client := wsClient{
		socket: c,
	}

	for {
		mt, message, err := c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			break
		}

		if mt != websocket.TextMessage {
			log.Println("discarding received binary message")
			continue
		}

		log.Printf("recv: %s", message)
		reply := decodeAndExecute(&client, string(message))

		if !client.write(reply) {
			break
		}
	}
}

func decodeAndExecute(client *wsClient, message string) string {
	parts := strings.Split(message, " ")

	switch parts[0] {
	case "register":
		if len(parts) < 3 {
			return "error Bad Request"
		}

		err := backend.Register(parts[1], parts[2])
		if err != nil {
			return err.Error()
		}

		client.username = parts[1]
		return "OK"

	case "login":
		if len(parts) < 3 {
			return "error Bad Request"
		}

		err := backend.Login(parts[1], parts[2], client)
		if err != nil {
			return err.Error()
		}

		client.username = parts[1]
		return "OK"

	case "post":
		if client.username == "" {
			return "error Unauthorized"
		}

		if len(parts) < 2 {
			return "error Bad Request"
		}

		msgID, err := backend.Post(client.username, strings.Join(parts[1:], " "))
		if err != nil {
			return err.Error()
		}

		return "OK " + strconv.FormatUint(msgID, 10)
	}

	return "error Bad Request"
}

func (client *wsClient) write(reply string) bool {
	log.Println("write:", reply)

	client.lock.Lock()
	err := client.socket.WriteMessage(websocket.TextMessage, []byte(reply))
	client.lock.Unlock()

	if err == nil {
		return true
	}

	log.Println("write:", err)
	return false
}

func (client *wsClient) Process(msg Message) {
	if msg.Poster.Username != client.username {
		mentioned := false
		for _, name := range msg.Mentions {
			if name == client.username {
				mentioned = true
				break
			}
		}
		if !mentioned {
			return
		}
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Println("failed to convert msg to JSON: ", msg.ID)
		return
	}

	client.write("buzz " + string(encoded))
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
	http.HandleFunc("/ws", accept)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(static))))
	http.Handle("/", http.RedirectHandler("/static/", http.StatusMovedPermanently))
	log.Fatal(http.ListenAndServe(endpoint, nil))
}
