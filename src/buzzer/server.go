package buzzer

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// MessageID is a unique identifier for a posted message.
type MessageID = uint64

// Message is a message posted by a user.
type Message struct {
	ID     MessageID `json:"id"`
	Text   string    `json:"text"`
	Poster *User     `json:"poster"`
	Posted time.Time `json:"posted"`
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
	Login(username, password string, client Client) error
	Logout(username string, client Client)
}

// Client implements sending a message from the server to the client and
// ensures it is done in a thread-safe way.
type Client interface {
	Process(msg Message)
}

// StartServer properly initializes, starts, and returns a new Server.
func StartServer() Server {
	actual := newBasicServer()
	server := newConcServer(actual)
	go server.process()
	return server
}

type concServer struct {
	actual                                                            *basicServer
	post, follow, unfollow, messages, tagged, register, login, logout chan request
	shutdown                                                          chan bool
}

func newConcServer(actual *basicServer) *concServer {
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
			err := server.actual.Login(req.args[0], req.args[1], req.client)
			go respond(&req, response{error: err})

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
	server.messages <- request{
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

func (server *concServer) Login(username, password string, client Client) error {
	resp := make(chan response)
	server.login <- request{
		args:   [2]string{username, password},
		client: client,
		resp:   resp,
	}
	reply := <-resp
	return reply.error
}

func (server *concServer) Logout(username string, client Client) {
	server.logout <- request{
		args:   [2]string{username, ""},
		client: client,
	}
}

// basicServer is a implementation of Server that can only be used serially.
type basicServer struct {
	lastID   MessageID
	messages map[MessageID]Message
	users    map[string]*User
	clients  []Client
}

func newBasicServer() *basicServer {
	return &basicServer{
		messages: make(map[MessageID]Message),
		users:    make(map[string]*User),
	}
}

// Post parses any mentions or tags then adds message to the list of messages.
func (server *basicServer) Post(username, message string) (MessageID, error) {
	user, ok := server.users[username]
	if !ok {
		return 0, errors.New("Unknown user")
	}

	server.lastID++

	msg := Message{
		ID:     server.lastID,
		Text:   message,
		Poster: user,
		Posted: time.Now(),
		// Mentions: parseMentions(message),
		// Tags: parseTags(messag)
	}

	server.messages[msg.ID] = msg

	go func(clients []Client) {
		for _, client := range clients {
			client.Process(msg)
		}
	}(server.clients)

	return msg.ID, nil
}

// Follow adds followee to follower's list of followers.
func (server *basicServer) Follow(followee, follower string) error {
	if followee == follower {
		return errors.New("Follower cannot follow themself")
	}

	ufollowee, followeeExists := server.users[followee]
	if !followeeExists {
		return errors.New("Unknown user: " + followee)
	}

	ufollower, followerExists := server.users[follower]
	if !followerExists {
		return errors.New("Unknown user: " + follower)
	}

	ufollower.follows[ufollowee] = true
	ufollowee.followers[ufollower] = true

	return nil
}

// Unfollow removes followee from follower's list of followers.
func (server *basicServer) Unfollow(followee, follower string) error {
	if followee == follower {
		return errors.New("Follower is not following themself")
	}

	ufollowee, followeeExists := server.users[followee]
	if !followeeExists {
		return errors.New("Unknown user: " + followee)
	}

	ufollower, followerExists := server.users[follower]
	if !followerExists {
		return errors.New("Unknown user: " + follower)
	}

	delete(ufollower.follows, ufollowee)
	delete(ufollowee.follows, ufollower)

	return nil
}

// Messages retrieves all posts made by a user and any posts in which username
// is mentioned using "@username".
func (server *basicServer) Messages(username string) []Message {
	var messages []Message

	user, ok := server.users[username]
	if !ok {
		return messages
	}

	for _, msg := range server.messages {
		if msg.Poster.Username == user.Username {
			messages = append(messages, msg)
		}
	}

	return messages
}

// Tagged retrieves all messages containing "#tag".
func (server *basicServer) Tagged(tag string) []Message {
	var messages []Message

	for _, msg := range server.messages {
		if strings.Contains(msg.Text, "#"+tag) {
			messages = append(messages, msg)
		}
	}

	return messages
}

var validUsernameRegex = regexp.MustCompile(`^\w+$`)

// Register checks that the username is available then files the username and
// password.
func (server *basicServer) Register(username, password string) error {
	if !validUsernameRegex.MatchString(username) {
		return errors.New("Invalid username")
	}

	if len(password) == 0 {
		return errors.New("Invalid password")
	}

	_, ok := server.users[username]
	if ok {
		return errors.New("Username taken")
	}

	server.users[username] = &User{
		Username:  username,
		password:  password,
		follows:   make(userSet),
		followers: make(userSet),
	}

	return nil
}

// Login verify the username and password with their known credentials.
func (server *basicServer) Login(username, password string, client Client) error {
	if !validUsernameRegex.MatchString(username) {
		return errors.New("Invalid username")
	}

	if len(password) == 0 {
		return errors.New("Invalid password")
	}

	user, ok := server.users[username]
	if !ok || user.password != password {
		return errors.New("Invalid credentials")
	}

	server.clients = append(server.clients, client)

	return nil
}

// Logout removes the username from the list of active clients; no further
// messages will be sent.
func (server *basicServer) Logout(username string, client Client) {
	_, ok := server.users[username]
	if !ok {
		return // User not found.
	}

	remaining := server.clients[:0]
	for _, c := range server.clients {
		if client == c {
			continue
		}
		remaining = append(remaining, c)
	}

	server.clients = remaining
}
