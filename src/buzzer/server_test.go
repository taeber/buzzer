package buzzer

import (
	"testing"
)

func TestPostByUnknownUserFails(t *testing.T) {
	var srv Server

	msgID, err := srv.Post("taeber", "Buzz! buzz!")
	if msgID != 0 || err == nil || err.Error() != "Unknown user" {
		t.Errorf("Post() is not validating username")
	}

}
