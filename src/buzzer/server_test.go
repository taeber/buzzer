package buzzer

import (
	"testing"
)

func BenchmarkChannelServerPost(b *testing.B) {
	basic := newKernel()
	basic.Register("tester", "testing")

	server := newChannelServer(basic)
	go server.process()

	for i := 0; i < b.N; i++ {
		server.Post("tester", "Buzzer message")
	}

	server.shutdown <- true
}
