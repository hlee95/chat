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
// - sender: recipient username
// - recipient: recipient username
// - messageType: one of "plaintext", "image_link", "video_link"
// - content: the text of the message
//
// Sample curl request:
// curl -d '{"sender":"user2", "recipient":"user1", "messageType":"plaintext", "content":"Hi there!"}' -H "Content-Type: application/json" -X POST localhost:18000/messages
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
  log.Printf("Received POST at /messages for sender %s and recipient %s", senderName, recipientName)
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
// - sender: sender username
// - recipient: recipient username
// - [messagesPerPage]: optional number of messages per page
// - [pageToLoad]: optional page number to show (0 indexed)
//
// Note that the order of the sender and recipient does not matter, they are
// simply better names than "username1" and "username 2"
//
// Sample curl request:
// curl "localhost:18000/messages?sender=user1&recipient=user2&messagesPerPage=2&pageToLoad=1"
func (server *ChatServer) fetchMessages(w http.ResponseWriter, r *http.Request) {
  // Parse request.
  u, err := url.Parse(r.URL.String())
  if err != nil {
    log.Printf("Couldn't parse GET at /messages, %+v", r)
    http.Error(w, "bad GET request at /messages, could not parse", http.StatusBadRequest)
    return
  }
  params := u.Query()
  if len(params["sender"]) != 1 || len(params["recipient"]) != 1 {
    log.Printf("Couldn't parse GET at /messages, %+v", r)
    http.Error(w, "bad GET request at /messages, could not parse", http.StatusBadRequest)
    return
  }
  // Confirm that parameters are all valid.
  senderName := params.Get("sender")
  recipientName := params.Get("recipient")
  _, haveMessagesPerPage := params["messagesPerPage"]
  _, havePageToLoad := params["pageToLoad"]
  if (haveMessagesPerPage && !havePageToLoad) ||
     (havePageToLoad && !haveMessagesPerPage) {
    log.Printf("Must provide both or neither of messagesPerPage and pageToLoad")
    http.Error(w, "must provide both or neither of messagesPerPage and pageToLoad", http.StatusBadRequest)
    return
  }
  if len(params["messagesPerPage"]) > 1 || len(params["pageToLoad"]) > 1 {
    log.Printf("Too many values for messagesPerPage or pageToLoad")
    http.Error(w, "too many values for messagesPerPage or pageToLoad", http.StatusBadRequest)
    return
  }
  log.Printf("Received GET at /messages for %s and %s", senderName, recipientName)
  // Get messages.
  messages, err := server.db.FetchMessages(senderName, recipientName)
  if err != nil {
    log.Printf("Error fetching messages from db: %s", err.Error())
    http.Error(w, "couldn't fetch messages", http.StatusInternalServerError)
    return
  }
  // Return only a subset of the messages if specified.
  if haveMessagesPerPage {
    messagesPerPage, err := strconv.Atoi(params["messagesPerPage"][0])
    if err != nil {
      log.Printf("Error parsing messagesPerPage")
      http.Error(w, "error parsing messagesPerPage", http.StatusBadRequest)
      return
    }
    pageToLoad, err := strconv.Atoi(params["pageToLoad"][0])
    if err != nil {
      log.Printf("Error parsing pageToLoad")
      http.Error(w, "error parsing pageToLoad", http.StatusBadRequest)
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
