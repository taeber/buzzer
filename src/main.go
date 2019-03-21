package main

import (
	"buzzer"
	"fmt"
	"os"
)

func main() {
	var srv buzzer.Server

	fmt.Println("Hello!")

	msgID, err := srv.Post("taeber", "Buzz! buzz!")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(msgID)
}
