package chatserver

import (
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net/http"
  "net/url"
  "strconv"
)

// Struct for decoding JSON body for POST requests at /messages.
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
// Note that we allow users to send messages to themselves.
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
  senderName, recipientName, usePagination, messagesPerPage, pageToLoad, err := server.parseFetchMessages(r)
  if (err != nil) {
    http.Error(w, fmt.Sprintf("bad GET request at /messages, could not parse, %s", err.Error()), http.StatusBadRequest)
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
  if usePagination {
    // Get the correct slice of messages.
    start := pageToLoad * messagesPerPage
    end := (pageToLoad + 1) * messagesPerPage
    // Check for boundary conditions.
    if len(messages) <= start || start < 0 || messagesPerPage <= 0 {
      log.Printf("Impossible pagination request, bad messagesPerPage and/or pageToLoad")
      http.Error(w, "bad messagesPerPage or pageToLoad, no results found for desired page", http.StatusBadRequest)
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

// Parse GET request for /messages.
// Helper function to increase readability.
// Returns parsed values or error.
func (server *ChatServer) parseFetchMessages(r *http.Request) (senderName string, recipientName string, usePagination bool, messagesPerPage int, pageToLoad int, err_ error) {
// Parse request.
  var err error
  u, err := url.Parse(r.URL.String())
  if err != nil {
    return "", "", false, -1, -1, errors.New("Couldn't parse GET at /messages")
  }
  params := u.Query()
  if len(params["sender"]) != 1 || len(params["recipient"]) != 1 {
    return "", "", false, -1, -1, errors.New("Couldn't parse GET at /messages")
  }
  senderName = params.Get("sender")
  recipientName = params.Get("recipient")
  // Check that messagesPerPage and pageToLoad either both have 1 value or
  // both have 0 values provided, and that they are parsable as integers.
  _, haveMessagesPerPage := params["messagesPerPage"]
  _, havePageToLoad := params["pageToLoad"]
  if (haveMessagesPerPage && !havePageToLoad) ||
     (havePageToLoad && !haveMessagesPerPage) ||
     len(params["messagesPerPage"]) > 1 ||
     len(params["pageToLoad"]) > 1{
    return "", "", false, -1, -1, errors.New("Expect messagesPerPage and pageToLoad to both have 0 or 1 values")
  }
  if (haveMessagesPerPage) {
    messagesPerPage, err = strconv.Atoi(params["messagesPerPage"][0])
    if err != nil {
      return "", "", false, -1, -1, errors.New("Error parsing messagesPerPage")
    }
    pageToLoad, err = strconv.Atoi(params["pageToLoad"][0])
    if err != nil {
      return "", "", false, -1, -1, errors.New("Error parsing pageToLoad")
    }
    return senderName, recipientName, true, messagesPerPage, pageToLoad, nil
  } else {
    return senderName, recipientName, false, -1, -1, nil
  }
}
