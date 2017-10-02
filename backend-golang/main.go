package main

import "app/chatserver"

// Entry point for our backend. Simply starts up the server.
func main() {
	server := new(chatserver.ChatServer)
	server.Start()
}
