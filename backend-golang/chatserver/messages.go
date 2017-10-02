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

// Request handler for /messages.
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
// - Sender: recipient username
// - Recipient: recipient username
// - MessageType: one of "plaintext", "image_link", "video_link"
// - Content: the text of the message
func (server *ChatServer) sendMessage(w http.ResponseWriter, r *http.Request) {
  // Parse request.
  var body sendMessageStruct
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&body); err != nil {
    log.Printf("Bad POST request received at /messages, %+v", r)
    http.Error(w, "bad POST request, could not parse", http.StatusBadRequest)
    return
  }
  senderName := body.Sender
  recipientName := body.Recipient
  messageType := body.MessageType
  content := body.Content
  // TODO: Can users send messages to themselves? I don't see why not so I'll allow it.
  w.Header().Add("Content-Type", "application/json")
  id, err := server.db.AddMessage(senderName, recipientName, messageType, content)
  if err != nil {
    log.Printf("error: %s", err.Error())
    http.Error(w, "couldn't sent message", http.StatusInternalServerError)
    return
  }
  // Success.
  log.Printf("Successfully stored message from %s to %s", senderName, recipientName)
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(map[string]string{
    "sender": senderName,
    "recipient": recipientName,
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
