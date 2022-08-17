package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"wb/tasks/lzero/db"
	"wb/tasks/lzero/nats"
	"wb/tasks/lzero/view"
)

// Application entry point
func main() {
	// Create all the stuff
	messageChannel := make(chan string, 256)

	testNats, err := nats.Create("test-cluster", "test-client", "test-channel", messageChannel)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Nats up")

	testDb, err := db.Create("postgresql://postgres:beijing@localhost/wb", messageChannel)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Db up")

	testView, err := view.Create(":3000", testDb)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("View up")

	// Stolen from https://stackoverflow.com/a/18158859
	interruptChannel := make(chan os.Signal)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)
	<-interruptChannel

	// Destroy all the stuff
	testView.Destroy()
	fmt.Println("View stopped")
	testDb.Destroy()
	fmt.Println("Db stopped")
	testNats.Destroy()
	fmt.Println("Nats stopped")
	close(messageChannel)
}
