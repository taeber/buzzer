package main

import (
	"bufio"
	"buzzer"
	"fmt"
	"os"
	"strings"
)

var srv *buzzer.Server

func main() {
	srv = buzzer.NewServer()

	if len(os.Args) > 1 && os.Args[1] == "-i" {
		shell()
		return
	}

	username := "taeber"
	password := "rapczak"
	srv.Register(username, password)
	srv.Register("bob", "ross")
	// srv.Login(username, password)
	srv.Post(username, "Hello, @world!")
	srv.Post(username, "Anyone from #cop5618sp19?")
	srv.Post(username, "@bob u there?")
	srv.Follow("taeber", "bob")
	srv.Post("bob", "Happy Trees")

	fmt.Println("BuzzFeed")
	for _, msg := range srv.Messages("taeber") {
		fmt.Printf(" %-10d\t%s\n", msg.ID, msg.Text)
	}

	fmt.Println("#cop5618sp19")
	for _, msg := range srv.Tagged("cop5618sp19") {
		fmt.Printf(" %-10d\t%s\n", msg.ID, msg.Text)
	}
}

func shell() {
	fmt.Println("Ready.")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := strings.Split(scanner.Text(), " ")
		if len(command) == 0 {
			continue
		}

		switch command[0] {
		case "reg":
			if len(command) == 3 {
				if err := srv.Register(command[1], command[2]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					fmt.Println("OK")
				}
				continue
			}

		case "post":
			if len(command) >= 3 {
				if msgID, err := srv.Post(command[1], strings.Join(command[2:], " ")); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					fmt.Println(msgID)
				}
				continue
			}

		case "feed":
			if len(command) == 2 {
				for _, msg := range srv.Messages(command[1]) {
					fmt.Printf("%10d\t%s\n", msg.ID, msg.Text)
				}
			}
			continue

		case "tag":
			if len(command) == 2 {
				for _, msg := range srv.Tagged(command[1]) {
					fmt.Printf("%10d\t%s\n", msg.ID, msg.Text)
				}
			}
			continue

		case "follow":
			if len(command) == 3 {
				if err := srv.Follow(command[1], command[2]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					fmt.Println("OK")
				}
			}
			continue

		case "unfollow":
			if len(command) == 3 {
				if err := srv.Unfollow(command[1], command[2]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					fmt.Println("OK")
				}
			}
			continue
		}

		fmt.Println("Invalid command or command arguments")
	}
}
