package buzzer

// MessageID is a unique identifier for a posted message.
type MessageID uint

// Message is a message posted by a user.
type Message struct {
	ID       MessageID
	Text     string
	Username string
}

// Server coordinates all activity for Buzzer. This is meant to be a low-level
// kernel of sorts that is wrapped by a protocol-specific handler, such as one
// for WebSockets. Specifically, authorization is assumed; no security checks
// are performed by these functions.
type Server struct {
	messages map[MessageID]Message
	users    map[string]string
}

// Post parses any mentions or tags then adds message to the list of messages.
func (*Server) Post(username, message string) (MessageID, error) {
	panic("Unimplemented")
}

// Follow adds followee to follower's list of followers.
func (*Server) Follow(followee, follower string) error {
	panic("Unimplemented")
}

// Unfollow removes followee from follower's list of followers.
func (*Server) Unfollow(followee, follower string) error {
	panic("Unimplemented")
}

// Messages retrieves all posts made by a user and any posts in which username
// is mentioned using "@username".
func (*Server) Messages(username string) []Message {
	panic("Unimplemented")
}

// Tagged retrieves all messages containing "#tag".
func (server *Server) Tagged(tag string) []Message {
	panic("Unimplemented")
}

// Register checks that the username is available then files the username and
// password.
func (*Server) Register(username, password string) error {
	panic("Unimplemented")
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
