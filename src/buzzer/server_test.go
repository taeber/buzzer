package buzzer

import (
	"testing"
)

func TestPostByUnknownUserFails(t *testing.T) {
	var srv basicServer

	msgID, err := srv.Post("taeber", "Buzz! buzz!")
	if msgID != 0 || err == nil || err.Error() != "Unknown user" {
		t.Errorf("Post() is not validating username")
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
