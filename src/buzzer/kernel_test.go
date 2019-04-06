package buzzer

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

func TestPostByUnknownUserFails(t *testing.T) {
	var srv kernel

	msgID, err := srv.Post("taeber", "Buzz! buzz!")
	if msgID != 0 || err == nil || err.Error() != "Unknown user" {
		t.Error("Post() is not validating username")
	}
}

// TestCopySlice ensures my understanding of slices allows for safe copying
// of User within the Login() function.
func TestCopySlice(t *testing.T) {
	type person struct {
		name    string
		friends []string
	}

	tom := person{"Tom", []string{"Jerry"}}
	tomRef := &tom
	tomCopy := *tomRef

	tom.friends = append(tom.friends, "Jane")

	if len(tomCopy.friends) != 1 {
		t.Error("You've misunderstood the copy semantics")
	}
	if len(tomRef.friends) != 2 {
		t.Error("You've misunderstood the pointer semantics")
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
	expected := `{"id":42,"text":"I do!","poster":{"username":"taeber"},"posted":"2012-06-23T13:30:00-04:00"}`
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

func BenchmarkKernelPost(b *testing.B) {
	basic := newKernel()
	basic.Register("tester", "testing")

	for i := 0; i < b.N; i++ {
		basic.Post("tester", "Buzzer message")
	}
}
