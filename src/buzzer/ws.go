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

	"github.com/gorilla/websocket"
)

type subscription struct {
	followee, follower string
	unfollow           bool
}

type wsClient struct {
	username  chan string // Alternative is to use sync/atomic.Value.
	socket    *websocket.Conn
	send      chan string
	subscribe chan subscription
}

var backend Server
var upgrader websocket.Upgrader

// accept handles a new HTTP connection by upgrading it to a WebSocket one. It
// also creates three goroutines: one reader for handling incoming messages
// one writer for sending messages, and one processer which decodes the
// messages, performs some action, then responds. Then, it waits around until
// it gets the shutdown signal.
func accept(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	client := wsClient{
		username:  make(chan string, 1),
		socket:    c,
		send:      make(chan string),
		subscribe: make(chan subscription),
	}

	// client.username is used as a semaphore of sorts. There can be multiple
	// goroutines reading and writing to it. Therefore, we set the initial
	// value here, then expect subsequent calls to aquire then release it.
	go func() {
		client.username <- ""
	}()

	received := make(chan string)
	shutdown := make(chan bool)

	// Handles all incoming messages.
	go func() {
		defer func() { shutdown <- true }()
		for {
			mt, msg, err := c.ReadMessage()

			if err != nil {
				log.Println("read:", err)
				break
			}

			if mt != websocket.TextMessage {
				log.Println("discarding received binary message")
				continue
			}

			log.Printf("recv: %s", msg)
			received <- string(msg)
		}
	}()

	// Sends any outgoing messages.
	go func() {
		defer func() { shutdown <- true }()
		for {
			msg := <-client.send
			log.Println("write:", msg)

			err := client.socket.WriteMessage(websocket.TextMessage, []byte(msg))
			if err == nil {
				continue
			}

			log.Println("write:", err)
			shutdown <- true
			break
		}
	}()

	// Processes any messages received.
	go func() {
		defer func() { shutdown <- true }()
		for {
			select {
			case msg := <-received:
				client.decodeAndExecute(msg)

			case sub := <-client.subscribe:
				username := client.getUsername()

				if sub.follower != username {
					continue
				}

				if sub.unfollow {
					client.Write("unfollow " + sub.followee)
				} else {
					client.Write("follow " + sub.followee)
				}
			}
		}
	}()

	// Wait for shutdown.
	<-shutdown

	username := <-client.username
	if username != "" {
		backend.Logout(username, &client)
	}
}

func (client *wsClient) getUsername() (username string) {
	username = <-client.username
	client.username <- username
	return
}

func (client *wsClient) setUsername(username string) {
	<-client.username
	client.username <- username
}

func (client *wsClient) decodeAndExecute(message string) {
	const (
		errBadRequest   = "error Bad Request"
		errUnauthorized = "error Unauthorized"
	)

	parts := strings.Split(message, " ")
	username := client.getUsername()

	switch parts[0] {
	case "register":
		if len(parts) < 3 {
			client.Write(errBadRequest)
			return
		}

		err := backend.Register(parts[1], parts[2])
		if err != nil {
			client.Write(err.Error())
			return
		}

		client.Write("OK")

	case "login":
		if len(parts) < 3 {
			client.Write(errBadRequest)
			return
		}

		if username != "" {
			backend.Logout(username, client)
		}

		user, err := backend.Login(parts[1], parts[2], client)
		if err != nil {
			client.Write(err.Error())
			return
		}

		client.setUsername(parts[1])
		client.Write("OK")

		for followee := range user.follows {
			client.Write("follow " + followee.Username)
		}

	case "post":
		if username == "" {
			client.Write(errUnauthorized)
			return
		}

		if len(parts) < 2 {
			client.Write(errBadRequest)
			return
		}

		msgID, err := backend.Post(username, strings.Join(parts[1:], " "))
		if err != nil {
			client.Write(err.Error())
			return
		}

		client.Write("OK " + strconv.FormatUint(msgID, 10))

	case "buzzfeed":
		if len(parts) < 2 {
			client.Write(errBadRequest)
			return
		}

		msgs := backend.Messages(parts[1])
		for _, msg := range msgs {
			encoded, err := json.Marshal(msg)
			if err != nil {
				log.Println("failed to convert msg to JSON: ", msg.ID)
				return
			}

			client.Write("buzz " + string(encoded))
		}

	case "follow":
		if username == "" {
			client.Write(errUnauthorized)
			return
		}

		if len(parts) < 2 {
			client.Write(errBadRequest)
			return
		}

		err := backend.Follow(parts[1], username)
		if err != nil {
			client.Write(err.Error())
			return
		}
		// client.Write("follow " + parts[1])

	case "unfollow":
		if username == "" {
			client.Write(errUnauthorized)
			return
		}

		if len(parts) < 2 {
			client.Write(errBadRequest)
			return
		}

		err := backend.Unfollow(parts[1], username)
		if err != nil {
			client.Write(err.Error())
			return
		}
		// client.Write("unfollow " + parts[1])

	default:
		client.Write(errBadRequest)
	}
}

func (client *wsClient) Process(msg Message) {
	username := client.getUsername()
	interested := msg.Poster.Username == username

	if !interested {
		// Check if user is mentioned.
		for _, name := range msg.Mentions {
			if name == username {
				interested = true
				break
			}
		}
	}

	if !interested {
		// Check if following poster.
		for follower := range msg.Poster.followers {
			if follower.Username == username {
				interested = true
				break
			}
		}
	}

	if !interested {
		return
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Println("failed to convert msg to JSON: ", msg.ID)
		return
	}

	client.send <- "buzz " + string(encoded)
}

func (client *wsClient) Subscription(followee, follower string, unfollow bool) {
	client.subscribe <- subscription{followee, follower, unfollow}
}

func (client *wsClient) Write(reply string) {
	client.send <- reply
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
