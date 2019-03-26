package buzzer

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

func TestPostByUnknownUserFails(t *testing.T) {
	var srv basicServer

	msgID, err := srv.Post("taeber", "Buzz! buzz!")
	if msgID != 0 || err == nil || err.Error() != "Unknown user" {
		t.Error("Post() is not validating username")
	}
}

func TestJSONMarshalling(t *testing.T) {
	user := User{"taeber", "secret", nil, nil}
	msg := Message{
		ID:     42,
		Text:   "I do!",
		Poster: &user,
		Posted: time.Date(2012, 06, 23, 13, 30, 0, 0, time.FixedZone("EST", -4*60*60)),
	}

	out, _ := json.Marshal(msg)
	expected := `{"ID":42,"Text":"I do!","Poster":{"Username":"taeber"},"Posted":"2012-06-23T13:30:00-04:00"}`
	if expected != string(out) {
		t.Errorf("JSON decoding failed:\n\t%s\n", out)
	}
}

func TestRegex(t *testing.T) {
	var validUsernameRegex = regexp.MustCompile(`^\w+$`)
	acceptable := []string{"therealrobboss", "yogi_bear", "_", "__tom____"}
	for _, name := range acceptable {
		if !validUsernameRegex.MatchString(name) {
			t.Errorf("Wrong regular expression; failed on:\t%s\n", name)
		}
	}

	unacceptable := []string{"", "Space the final ", "...", "first.last"}
	for _, name := range unacceptable {
		if validUsernameRegex.MatchString(name) {
			t.Errorf("Wrong regular expression; allowed:\t%s\n", name)
		}
	}

}

func BenchmarkBasicServerPost(b *testing.B) {
	basic := newBasicServer()
	basic.Register("tester", "testing")

	for i := 0; i < b.N; i++ {
		basic.Post("tester", "Buzzer message")
	}
}

func BenchmarkConcServerPost(b *testing.B) {
	basic := newBasicServer()
	basic.Register("tester", "testing")

	server := newConcServer(basic)
	go server.process()

	for i := 0; i < b.N; i++ {
		server.Post("tester", "Buzzer message")
	}

	server.shutdown <- true
}
