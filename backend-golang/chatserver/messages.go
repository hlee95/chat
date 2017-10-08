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
  Sender      string
  Recipient   string
  MessageType string
  Content     string
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
// - sender: sender username
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
  senderName, recipientName, messageType, content, err := server.parseSendMessage(r)
  if err != nil {
    http.Error(w,fmt.Sprintf(
      "bad POST request at /messages, couldn't parse, error: %s",
      err.Error()),
    http.StatusBadRequest)
    return
  }

  log.Printf("Received POST at /messages for sender %s and recipient %s", senderName, recipientName)
  id, err := server.db.AddMessage(senderName, recipientName, messageType, content)
  if err != nil {
    log.Printf("Error adding message to db: %s", err.Error())
    http.Error(w, fmt.Sprintf("Couldn't send message: %s", err.Error()), http.StatusInternalServerError)
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

// Parse POST request for /messages.
// Returns parsed values or error.
func (server *ChatServer) parseSendMessage(r *http.Request) (senderName string, recipientName string, messageType string, content string, err error) {
  var body sendMessageStruct
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&body); err != nil {
    return "", "", "", "", errors.New("couldn't decode JSON")
  }
  senderName = body.Sender
  recipientName = body.Recipient
  messageType = body.MessageType
  content = body.Content
  // Ignore empty messages.
  if len(content) <= 0 {
    return "", "", "", "", errors.New(fmt.Sprintf("rejecting empty message"))
  }
  if messageType != MESSAGE_TYPE_PLAINTEXT && messageType != MESSAGE_TYPE_IMAGE_LINK &&
     messageType != MESSAGE_TYPE_VIDEO_LINK {
      return "", "", "", "", errors.New(fmt.Sprintf("invalid messageType %s", messageType))
  }
  return senderName, recipientName, messageType, content, nil
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
  fetchMessagesParams, err := server.parseFetchMessages(r)
  if (err != nil) {
    http.Error(w, fmt.Sprintf("bad GET request at /messages, could not parse, %s", err.Error()), http.StatusBadRequest)
    return
  }
  log.Printf("Received GET at /messages for %s and %s", fetchMessagesParams.senderName,
                                                        fetchMessagesParams.recipientName)
  // Get messages.
  messages, err := server.db.FetchMessages(fetchMessagesParams)
  if err != nil {
    log.Printf("Error fetching messages from db: %s", err.Error())
    http.Error(w, fmt.Sprintf("Couldn't fetch messages: %s", err.Error()), http.StatusInternalServerError)
    return
  }
  // Try to send response.
  log.Printf("Successfully fetched messages between %s and %s",
             fetchMessagesParams.senderName, fetchMessagesParams.recipientName)
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(messages); err != nil {
    log.Printf("Error formatting http response, %s", err.Error())
    http.Error(w, "error generating response", http.StatusInternalServerError)
  }
}

// Parse GET request for /messages.
// Returns parsed values or error.
func (server *ChatServer) parseFetchMessages(r *http.Request) (fetchMessagesParams *FetchMessagesParams, err error) {
  fetchMessagesParams = &FetchMessagesParams{}
  // Parse request.
  u, err := url.Parse(r.URL.String())
  if err != nil {
    err = errors.New("Couldn't parse GET at /messages")
    return
  }
  params := u.Query()
  if len(params["sender"]) != 1 || len(params["recipient"]) != 1 {
    err = errors.New("Couldn't parse GET at /messages")
    return
  }
  fetchMessagesParams.senderName = params.Get("sender")
  fetchMessagesParams.recipientName = params.Get("recipient")
  // Check that messagesPerPage and pageToLoad either both have 1 value or
  // both have 0 values provided, and that they are parsable as integers.
  _, haveMessagesPerPage := params["messagesPerPage"]
  _, havePageToLoad := params["pageToLoad"]
  if (haveMessagesPerPage && !havePageToLoad) ||
     (havePageToLoad && !haveMessagesPerPage) ||
     len(params["messagesPerPage"]) > 1 ||
     len(params["pageToLoad"]) > 1{
    err = errors.New("Expect messagesPerPage and pageToLoad to both have 0 or 1 values")
    return
  }
  if (haveMessagesPerPage) {
    fetchMessagesParams.usePagination = true
    fetchMessagesParams.messagesPerPage, err = strconv.Atoi(params["messagesPerPage"][0])
    if err != nil {
      err = errors.New("Error parsing messagesPerPage")
      return
    }
    fetchMessagesParams.pageToLoad, err = strconv.Atoi(params["pageToLoad"][0])
    if err != nil {
      err = errors.New("Error parsing pageToLoad")
      return
    }
  }
  return
}
