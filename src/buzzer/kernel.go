package buzzer

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// kernel is a implementation of Server that can only be used serially.
type kernel struct {
	lastID   MessageID
	messages map[MessageID]Message
	users    map[string]*User
	clients  []Client
}

func newKernel() *kernel {
	return &kernel{
		messages: make(map[MessageID]Message),
		users:    make(map[string]*User),
	}
}

// Post parses any mentions or tags then adds message to the list of messages.
func (server *kernel) Post(username, message string) (MessageID, error) {
	user, ok := server.users[username]
	if !ok {
		return 0, errors.New("Unknown user")
	}

	server.lastID++

	msg := Message{
		ID:       server.lastID,
		Text:     message,
		Poster:   user,
		Posted:   time.Now(),
		Mentions: parseMentions(message),
		Tags:     parseTags(message),
	}

	server.messages[msg.ID] = msg

	// WARNING: this creates a shallow copy of User. This is thread-safe
	// because slices in go are references and, in this case, point to
	// effectively immutable objects.
	snapshot := msg
	snapshot.Poster = &*user

	go func(clients []Client) {
		for _, client := range clients {
			client.Process(snapshot)
		}
	}(server.clients)

	return msg.ID, nil
}

// Follow adds followee to follower's list of followers.
func (server *kernel) Follow(followee, follower string) error {
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

	go func(clients []Client) {
		for _, client := range clients {
			client.Subscription(followee, follower, false)
		}
	}(server.clients)

	return nil
}

// Unfollow removes followee from follower's list of followers.
func (server *kernel) Unfollow(followee, follower string) error {
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
	delete(ufollowee.followers, ufollower)

	go func(clients []Client) {
		for _, client := range clients {
			client.Subscription(followee, follower, true)
		}
	}(server.clients)

	return nil
}

// Messages retrieves all posts made by a user and any posts in which username
// is mentioned using "@username".
func (server *kernel) Messages(username string) []Message {
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
func (server *kernel) Tagged(tag string) []Message {
	var messages []Message

	tag = strings.ToLower(tag)

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
func (server *kernel) Register(username, password string) error {
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
func (server *kernel) Login(username, password string, client Client) (*User, error) {
	if !validUsernameRegex.MatchString(username) {
		return nil, errors.New("Invalid username")
	}

	user, ok := server.users[username]
	if !ok {
		return nil, errors.New("Unknown user")
	}

	if user.password != password {
		return nil, errors.New("Invalid credentials")
	}

	server.clients = append(server.clients, client)

	// WARNING: this creates a shallow copy of User. This is thread-safe
	// because slices in go are references and, in this case, point to
	// effectively immutable objects.
	snapshot := *user

	return &snapshot, nil
}

// Logout removes the username from the list of active clients; no further
// messages will be sent.
func (server *kernel) Logout(username string, client Client) {
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
