package buzzer

import (
	"time"
)

// MessageID is a unique identifier for a posted message.
type MessageID = uint64

// Message is a message posted by a user.
type Message struct {
	ID       MessageID `json:"id"`
	Text     string    `json:"text"`
	Poster   *User     `json:"poster"`
	Posted   time.Time `json:"posted"`
	Mentions []string  `json:"mentions,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
}

// User is a person or bot that uses the service.
type User struct {
	Username  string `json:"username"`
	password  string
	follows   userSet
	followers userSet
}

// userSet is a set of unique users.
type userSet = map[*User]bool

// Server coordinates all activity for Buzzer. This is meant to be a low-level
// kernel of sorts that is wrapped by a protocol-specific handler, such as one
// for WebSockets. Specifically, authorization is assumed; no security checks
// are performed by these functions.
type Server interface {
	Post(username, message string) (MessageID, error)
	Follow(followee, follower string) error
	Unfollow(followee, follower string) error
	Messages(username string) []Message
	Tagged(tag string) []Message

	Register(username, password string) error
	Login(username, password string, client Client) (*User, error)
	Logout(username string, client Client)
}

// Client implements sending a message from the server to the client and
// ensures it is done in a thread-safe way.
type Client interface {
	Process(msg Message)
	Subscription(followee, follower string, unfollow bool)
}

// StartServer properly initializes, starts, and returns a new Server.
func StartServer() Server {
	actual := newKernel()
	server := newChannelServer(actual)
	go server.process()
	return server
}

type concServer struct {
	actual                                                            *kernel
	post, follow, unfollow, messages, tagged, register, login, logout chan request
	shutdown                                                          chan bool
}

func newChannelServer(actual *kernel) *concServer {
	return &concServer{
		actual:   actual,
		post:     make(chan request, 100),
		follow:   make(chan request, 100),
		unfollow: make(chan request, 100),
		messages: make(chan request, 100),
		tagged:   make(chan request, 100),
		register: make(chan request, 100),
		login:    make(chan request, 100),
		logout:   make(chan request, 100),
		shutdown: make(chan bool),
	}
}

type response struct {
	data interface{}
	error
}

type request struct {
	args   [2]string
	client Client
	resp   chan response
}

// process checks for a request in on any of the channels then forwards it to
// the serial version of the Server. Responses are given asynchronously, using
// a goroutine, to prevent blocking. This method ensures safe, concurrent
// access to the underlying data.
func (server *concServer) process() {
	for {
		select {
		case req := <-server.post:
			msgID, err := server.actual.Post(req.args[0], req.args[1])
			go respond(&req, response{data: msgID, error: err})

		case req := <-server.follow:
			err := server.actual.Follow(req.args[0], req.args[1])
			go respond(&req, response{error: err})

		case req := <-server.unfollow:
			err := server.actual.Unfollow(req.args[0], req.args[1])
			go respond(&req, response{error: err})

		case req := <-server.messages:
			msgs := server.actual.Messages(req.args[0])
			go respond(&req, response{data: msgs})

		case req := <-server.tagged:
			msgs := server.actual.Tagged(req.args[0])
			go respond(&req, response{data: msgs})

		case req := <-server.register:
			err := server.actual.Register(req.args[0], req.args[1])
			go respond(&req, response{error: err})

		case req := <-server.login:
			user, err := server.actual.Login(req.args[0], req.args[1], req.client)
			go respond(&req, response{data: user, error: err})

		case req := <-server.logout:
			server.actual.Logout(req.args[0], req.client)

		case <-server.shutdown:
			//TODO: what happens to items in the buffered channel? Do I need to empty them out and close all channels?
			return
		}
	}
}

func respond(req *request, res response) {
	req.resp <- res
}

func (server *concServer) Post(username, message string) (MessageID, error) {
	resp := make(chan response)
	server.post <- request{
		args: [2]string{username, message},
		resp: resp,
	}
	reply := <-resp
	return reply.data.(MessageID), reply.error
}

func (server *concServer) Follow(followee, follower string) error {
	resp := make(chan response)
	server.follow <- request{
		args: [2]string{followee, follower},
		resp: resp,
	}
	reply := <-resp
	return reply.error
}

func (server *concServer) Unfollow(followee, follower string) error {
	resp := make(chan response)
	server.unfollow <- request{
		args: [2]string{followee, follower},
		resp: resp,
	}
	reply := <-resp
	return reply.error
}

func (server *concServer) Messages(username string) []Message {
	resp := make(chan response)
	server.messages <- request{
		args: [2]string{username},
		resp: resp,
	}
	reply := <-resp
	return reply.data.([]Message)
}

func (server *concServer) Tagged(tag string) []Message {
	resp := make(chan response)
	server.tagged <- request{
		args: [2]string{tag},
		resp: resp,
	}
	reply := <-resp
	return reply.data.([]Message)
}

func (server *concServer) Register(username, password string) error {
	resp := make(chan response)
	server.register <- request{
		args: [2]string{username, password},
		resp: resp,
	}
	reply := <-resp
	return reply.error
}

func (server *concServer) Login(username, password string, client Client) (*User, error) {
	resp := make(chan response)
	server.login <- request{
		args:   [2]string{username, password},
		client: client,
		resp:   resp,
	}
	reply := <-resp
	return reply.data.(*User), reply.error
}

func (server *concServer) Logout(username string, client Client) {
	server.logout <- request{
		args:   [2]string{username, ""},
		client: client,
	}
}
