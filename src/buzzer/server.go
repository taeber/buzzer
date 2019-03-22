package buzzer

import (
	"errors"
	"strings"
)

// MessageID is a unique identifier for a posted message.
type MessageID uint

// Message is a message posted by a user.
type Message struct {
	ID   MessageID
	Text string
	User *User
}

// User is a person or bot that uses the service.
type User struct {
	Username  string
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
type Server struct {
	lastID   MessageID
	messages map[MessageID]Message
	users    map[string]*User
}

// NewServer properly initializes and returns a new Server.
func NewServer() *Server {
	return &Server{
		messages: make(map[MessageID]Message),
		users:    make(map[string]*User),
	}
}

// Post parses any mentions or tags then adds message to the list of messages.
func (server *Server) Post(username, message string) (MessageID, error) {
	user, ok := server.users[username]
	if !ok {
		return 0, errors.New("Unknown user")
	}

	server.lastID++

	msg := Message{
		ID:   server.lastID,
		Text: message,
		User: user,
		// Mentions: parseMentions(message),
		// Tags: parseTags(messag)
	}

	server.messages[msg.ID] = msg

	return msg.ID, nil
}

// Follow adds followee to follower's list of followers.
func (server *Server) Follow(followee, follower string) error {
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
func (server *Server) Unfollow(followee, follower string) error {
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
func (server *Server) Messages(username string) []Message {
	var messages []Message

	user, ok := server.users[username]
	if !ok {
		return messages
	}

	for _, msg := range server.messages {
		if msg.User.Username == user.Username {
			messages = append(messages, msg)
		}
	}

	return messages
}

// Tagged retrieves all messages containing "#tag".
func (server *Server) Tagged(tag string) []Message {
	var messages []Message

	for _, msg := range server.messages {
		if strings.Contains(msg.Text, "#"+tag) {
			messages = append(messages, msg)
		}
	}

	return messages
}

// Register checks that the username is available then files the username and
// password.
func (server *Server) Register(username, password string) error {
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
func (*Server) Login(username, password string) error {
	panic("Unimplemented")
}

// Logout removes the username from the list of active clients; no further
// messages will be sent.
func (*Server) Logout(username string) {
	panic("Unimplemented")
}
