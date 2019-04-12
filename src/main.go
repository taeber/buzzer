package main

import (
	"bufio"
	"buzzer"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var srv buzzer.Server

var endpoint = flag.String("addr", "0.0.0.0:8080", "http service address")
var interactive = flag.Bool("client", false, "Run in client/interactive mode")
var numActors = flag.Int("actors", 0, "Run with fake actors")

// There are two primary modes: interactive and non-interactive. Interactive
// allows the user to test the implementation of functions one at a time. The
// non-interactive mode starts a number of autonomous actors who continuously
// make random choices about what to do.
func main() {
	srv = buzzer.StartServer()
	flag.Usage = func() {
		fmt.Printf("Usage: %s [FLAGS] [WWWROOT]\n", os.Args[0])
		fmt.Println("  WWWROOT:\tpath to the Web client")
		fmt.Println("  FLAGS  :\t")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *interactive {
		shell()
		return
	}

	for i := 0; i < *numActors; i++ {
		go actor("user" + strconv.Itoa(i))
	}

	buzzer.StartWebServer(srv, *endpoint, flag.Arg(0))
}

func actor(name string) {
	if err := srv.Register(name, "Password? We don't need no stinkin' password!"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if _, err := srv.Post(name, "#user @"+name+" registered"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(name, ": registered.")
	for {
		choice := rand.Intn(3)
		switch choice {
		case 0:
			msg := fmt.Sprintf("I picked a random number: %d!", rand.Int())
			if msgID, err := srv.Post(name, msg); err == nil {
				fmt.Printf("%s: posted message %d: %s\n", name, msgID, msg)
			}

		case 1:
			other := fmt.Sprintf("user%d", rand.Intn(*numActors))
			if err := srv.Follow(name, other); err == nil {
				fmt.Printf("%s: started following %s\n", name, other)
			}

		case 2:
			other := fmt.Sprintf("user%d", rand.Intn(*numActors))
			if err := srv.Unfollow(name, other); err == nil {
				fmt.Printf("%s: stopped following %s\n", name, other)
			}
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
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
