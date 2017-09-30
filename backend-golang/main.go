package main

// Entry point for our backend. Simply starts up the server.
func main() {
	server := new(ChatServer)
	server.Start()
}
