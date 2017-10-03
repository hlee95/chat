package chatserver

import (
  "encoding/json"
  "log"
  "net/http"
  "net/url"
  "strconv"
)

type sendMessageStruct struct {
  Sender      string;
  Recipient   string;
  MessageType string;
  Content     string;
}

// Request handler for /messages.
func (server *ChatServer) handleMessages(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "application/json")
  switch r.Method {
  case http.MethodGet:
    server.fetchMessages(w, r)
  case http.MethodPost:
    server.sendMessage(w, r)
  default:
    // Unhandled request, respond with StatusMethodNotAllowed (405).
    log.Printf("Unknown request received at /messages, %+v", r)
    http.Error(w, "only GET and POST requests are accepted", http.StatusMethodNotAllowed)
  }
}

// Adds a message to the database.
// Expects a POST to /messages with the following parameters in the body:
// - Sender: recipient username
// - Recipient: recipient username
// - MessageType: one of "plaintext", "image_link", "video_link"
// - Content: the text of the message
//
// Sample curl request:
// curl -d '{"Sender":"user2", "Recipient":"user1", "MessageType":"plaintext", "Content":"Hi there!"}' -H "Content-Type: application/json" -X POST localhost:18000/messages
func (server *ChatServer) sendMessage(w http.ResponseWriter, r *http.Request) {
  // Parse request.
  var body sendMessageStruct
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&body); err != nil {
    log.Printf("Bad POST request received at /messages, %+v", r)
    http.Error(w, "bad POST request at /messages, could not parse", http.StatusBadRequest)
    return
  }
  senderName := body.Sender
  recipientName := body.Recipient
  messageType := body.MessageType
  content := body.Content
  // TODO: Can users send messages to themselves? I don't see why not so I'll allow it.
  id, err := server.db.AddMessage(senderName, recipientName, messageType, content)
  if err != nil {
    log.Printf("Error adding message to db: %s", err.Error())
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
// Expects a GET to /messages with the following query parameters:
// - Sender: sender username
// - Recipient: recipient username
// - [MessagesPerPage]: optional number of messages per page
// - [PageToLoad]: optional page number to show (0 indexed)
//
// Note that the order of the sender and recipient does not matter, they are
// simply better names than "username1" and "username 2"
//
// Sample curl request:
// curl "localhost:18000/messages?Sender=user1&Recipient=user2&MessagesPerPage=2&PageToLoad=1"
func (server *ChatServer) fetchMessages(w http.ResponseWriter, r *http.Request) {
  // Parse request.
  u, err := url.Parse(r.URL.String())
  if err != nil {
    log.Printf("Couldn't parse GET at /messages, %+v", r)
    http.Error(w, "bad GET request at /messages, could not parse", http.StatusBadRequest)
    return
  }
  params := u.Query()
  if len(params["Sender"]) != 1 || len(params["Recipient"]) != 1 {
    log.Printf("Couldn't parse GET at /messages, %+v", r)
    http.Error(w, "bad GET request at /messages, could not parse", http.StatusBadRequest)
    return
  }
  // Confirm that parameters are all valid.
  senderName := params.Get("Sender")
  recipientName := params.Get("Recipient")
  _, haveMessagesPerPage := params["MessagesPerPage"]
  _, havePageToLoad := params["PageToLoad"]
  if (haveMessagesPerPage && !havePageToLoad) ||
     (havePageToLoad && !haveMessagesPerPage) {
    log.Printf("Must provide both or neither of MessagesPerPage and PageToLoad")
    http.Error(w, "must provide both or neither of MessagesPerPage and PageToLoad", http.StatusBadRequest)
    return
  }
  if len(params["MessagesPerPage"]) > 1 || len(params["PageToLoad"]) > 1 {
    log.Printf("Too many values for MessagesPerPage or PageToLoad")
    http.Error(w, "too many values for MessagesPerPage or PageToLoad", http.StatusBadRequest)
    return
  }
  // Get messages.
  messages, err := server.db.FetchMessages(senderName, recipientName)
  if err != nil {
    log.Printf("Error fetching messages from db: %s", err.Error())
    http.Error(w, "couldn't fetch messages", http.StatusInternalServerError)
    return
  }
  // Return only a subset of the messages if specified.
  if haveMessagesPerPage {
    messagesPerPage, err := strconv.Atoi(params["MessagesPerPage"][0])
    if err != nil {
      log.Printf("Error parsing MessagesPerPage")
      http.Error(w, "error parsing MessagesPerPage", http.StatusBadRequest)
      return
    }
    pageToLoad, err := strconv.Atoi(params["PageToLoad"][0])
    if err != nil {
      log.Printf("Error parsing PageToLoad")
      http.Error(w, "error parsing PageToLoad", http.StatusBadRequest)
      return
    }
    // Get the correct slice of messages.
    start := pageToLoad * messagesPerPage
    end := (pageToLoad + 1) * messagesPerPage
    // Check for boundary conditions.
    if len(messages) <= start || start < 0 || messagesPerPage <= 0 {
      log.Printf("Impossible pagination request, bad MessagesPerPage and/or PageToLoad")
      http.Error(w, "desired MessagesPerPage and PageToLoad results in impossible page", http.StatusBadRequest)
      return
    }
    if len(messages) >= end {
      messages = messages[start:end]
    } else {
      messages = messages[start:]
    }
  }
  // Try to send response.
  log.Printf("Successfully fetched messages between %s and %s", senderName, recipientName)
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(messages); err != nil {
    log.Printf("Error formatting http response, %s", err.Error())
    http.Error(w, "error generating response", http.StatusInternalServerError)
  }
}
