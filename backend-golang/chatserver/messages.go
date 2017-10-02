package chatserver

import (
  "encoding/json"
  "log"
  "net/http"
  "strconv"
)

type sendMessageStruct struct {
  Sender      string;
  Recipient   string;
  MessageType string;
  Content     string;
}

type fetchMessagesStruct struct {
  Sender          string;
  Recipient       string;
  MessagesPerPage int;
  PageToLoad      int;
}

// Request handler for /users.
// Expects a POST request with the following parameters in the body
// - Username : maximum 10 characters
// - Password : maximum 72 characters (due to bcrypt limitation)
// Expects data in JSON, because it's easier to send JSON than url-encoded
// key value pairs in React, and our frontend is in React.
//
// Sample curl request:
// curl -d '{"Username":"hanna", "Password":"secret"}' -H "Content-Type: application/json" -X POST localhost:18000/users
func (server *ChatServer) handleMessages(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case http.MethodGet:
    server.fetchMessages(w, r)
  case http.MethodPost:
    server.sendMessage(w, r)
  default:
    // Unhandled request, respond with StatusMethodNotAllowed (405).
    log.Printf("Unknown request received at /messages, %+v", r)
    w.Header().Add("Content-Type", "application/json")
    http.Error(w, "only GET and POST requests are accepted", http.StatusMethodNotAllowed)
  }
}

// Adds a message to the database.
// Expects the following parameters:
// - sender username
// - recipient username
// - message type
// - message content
func (server *ChatServer) sendMessage(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "application/json")
  id, err := server.db.AddMessage("user1", "user2", "plaintext", "hello world")
  if err != nil {
    log.Printf("error: %s", err.Error())
    http.Error(w, "couldn't sent message", http.StatusInternalServerError)
    return
  }
  // log.Printf("Successfully stored message from %s to %s", senderName, recipientName)
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(map[string]string{
    "sender": "user1",
    "recipient": "user2",
    "message_id": strconv.FormatInt(id, 10),
  }); err != nil {
    log.Printf("Error formatting http response, %s", err.Error())
    http.Error(w, "error generating response", http.StatusInternalServerError)
  }
}

// Fetches messages between two users.
// Expects the following parameters:
// - sender username
// - recipient username
// - messages per page
// - page number
func (server *ChatServer) fetchMessages(w http.ResponseWriter, r *http.Request) {
  // TODO
}
