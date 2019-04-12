.PHONY: buzzer test benchmark cover run zip

PWD=$(shell pwd)
GOPATH=GOPATH=$(PWD)/lib:$(PWD)

buzzer:
	$(GOPATH) go get github.com/gorilla/websocket
	$(GOPATH) go build -o $@ src/main.go

test:
	$(GOPATH) go test buzzer

benchmark:
	$(GOPATH) go test -benchmem -run=^$$ buzzer -bench .

cover:
	$(GOPATH) go test -coverprofile cover.out buzzer
	$(GOPATH) go tool cover -html=cover.out -o cover.html
	open cover.html || echo "Open cover.html in your browser"

run:
	$(GOPATH) go run src/main.go src/client

zip:
	zip -r buzzer.zip src docs README.md

clean:
	rm -f cover.out cover.html buzzer
