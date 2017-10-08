package chatserver

import (
  "database/sql"
  "errors"
  "fmt"
  "log"
  _ "github.com/go-sql-driver/mysql"
)

// MySQL queries and statements.
const INSERT_USER = "INSERT INTO users(username, hash) VALUES(?, ?)"
const INSERT_MESSAGE = "INSERT INTO messages(sender_id, recipient_id, message_type, message_content, message_metadata_id) VALUES (?, ?, ?, ?, ?)"
const INSERT_MESSAGE_WITH_NO_METADATA = "INSERT INTO messages(sender_id, recipient_id, message_type, message_content) VALUES (?, ?, ?, ?)"
const INSERT_MESSAGES_IMAGE_METADATA = "INSERT INTO messages_metadata(width, height) VALUES(?, ?)"
const INSERT_MESSAGES_VIDEO_METADATA = "INSERT INTO messages_metadata(length, source) VALUES(?, ?)"

const SELECT_ID_FROM_USERNAME = "SELECT id FROM users WHERE username=?"
const SELECT_USERNAME_FROM_ID = "SELECT username FROM users WHERE id=?"
const SELECT_IMAGE_METADATA = "SELECT width, height FROM messages_metadata WHERE id=?"
const SELECT_VIDEO_METADATA = "SELECT length, source FROM messages_metadata WHERE id=?"
// Selects from messages and joins on the metadata_id if possible.
const SELECT_MESSAGES_BETWEEN_USERS = `SELECT messages.sender_id, messages.recipient_id, messages.message_type, messages.message_content, ` +
                                        `messages_metadata.width, messages_metadata.height, messages_metadata.length, messages_metadata.source ` +
                                      `FROM messages ` +
                                      `LEFT JOIN messages_metadata ON messages_metadata.id=messages.message_metadata_id ` +
                                      `WHERE (messages.sender_id=? AND messages.recipient_id=?) OR (messages.sender_id=? AND messages.recipient_id=?) ` +
                                      `ORDER BY messages.id `
const SELECT_MESSAGES_BETWEEN_USERS_WITH_LIMIT = SELECT_MESSAGES_BETWEEN_USERS +
                                                 `LIMIT ?, ?`
const SELECT_USER_CREDENTIALS = "SELECT hash FROM users WHERE username=?"



// ChatSQLClient wraps a connection to the database, and provides an
// api to the server.
// This abstraction is in case we want to swap out a different db
// for the project without affecting the logic in the server.
//
// API exposed to server includes the following:
// - NewChatSqlClient()
// - client.CreateUser(username)
// - client.CheckUserExists(username)
// - client.GetUserCredentials(username)
// - client.FetchMessages(senderName, recipientName)
// - client.AddMessage(senderName, recipientName, messageType, messageContent)
//
// ** Note that the server is responsible for handling errors propagated
// up by the db client. **
type ChatSQLClient struct {
  db *sql.DB
}

// Given a user, get its id.
func (client *ChatSQLClient) getUserId(username string) (int64, error) {
  var id int64
  err := client.db.QueryRow(SELECT_ID_FROM_USERNAME, username).Scan(&id)
  return id, err
}

// Create a new user in the database with the given username, password hash, and salt.
// Returns the id of the newly created user, or an error.
func (client *ChatSQLClient) CreateUser(username string, hash []byte) (id int64, err error) {
  res, err := client.db.Exec(INSERT_USER, username, hash)
  if err != nil {
    return -1, err
  }
  return res.LastInsertId()
}

// Retrieves the password hash and salt for the given username.
func (client *ChatSQLClient) GetUserCredentials(username string) (hash []byte, err error) {
  err = client.db.QueryRow(SELECT_USER_CREDENTIALS, username).Scan(&hash)
  return
}

// Adds a new message to the database. Returns the id of that message, or an error.
func (client *ChatSQLClient) AddMessage(senderName string, recipientName string, messageType string, content string) (id int64, err_ error) {
  // Find the associated ids of the two users.
  var err error
  senderId, err := client.getUserId(senderName)
  if err != nil {
    return -1, errors.New(fmt.Sprintf("no such user %s", senderName))
  }
  recipientId, err := client.getUserId(recipientName)
  if err != nil {
    return -1, errors.New(fmt.Sprintf("no such user %s", recipientName))
  }
  switch messageType {
  case MESSAGE_TYPE_PLAINTEXT:
    // For regular messages, insert without any metadata.
    res, err := client.db.Exec(INSERT_MESSAGE_WITH_NO_METADATA, senderId,
                               recipientId, messageType, content)
    if err != nil {
      return -1, err
    }
    return res.LastInsertId()
  case MESSAGE_TYPE_IMAGE_LINK, MESSAGE_TYPE_VIDEO_LINK:
    // TODO: Use a prepared statement.
    tx, err := client.db.Begin()
    var res sql.Result
    // First insert the metadata.
    if messageType == MESSAGE_TYPE_IMAGE_LINK {
      res, err = tx.Exec(INSERT_MESSAGES_IMAGE_METADATA, IMAGE_WIDTH,
                                IMAGE_HEIGHT)
    } else {
      res, err = tx.Exec(INSERT_MESSAGES_VIDEO_METADATA, VIDEO_LENGTH,
                                VIDEO_SOURCE)
    }
    if err != nil {
      tx.Rollback()
      return -1, err
    }
    metadataId, err := res.LastInsertId()
    if err != nil {
      tx.Rollback()
      return -1, err
    }
    // Then insert the message.
    res, err = tx.Exec(INSERT_MESSAGE, senderId, recipientId,
                              messageType, content, metadataId)
    if err != nil {
      tx.Rollback()
      return -1, err
    }
    if err = tx.Commit(); err != nil {
      tx.Rollback()
      return -1, err
    }
    return res.LastInsertId()
  default:
    return -1, errors.New(fmt.Sprintf("Unknown message type %s", messageType))
  }
}

// Gets messages between two users.
// Return an array of pointers to the Message struct.
func (client *ChatSQLClient) FetchMessages(params *FetchMessagesParams) (messages []*Message, err error) {
  // Find the associated ids of the two users.
  requestedSenderId, err := client.getUserId(params.senderName)
  if err != nil {
    err = errors.New(fmt.Sprintf("no such user %s", params.senderName))
    return nil, err
  }
  requestedRecipientId, err := client.getUserId(params.recipientName)
  if err != nil {
    err = errors.New(fmt.Sprintf("no such user %s", params.recipientName))
    return nil, err
  }
  // Get all rows, limit the number of entries depending on pagination.
  var senderId int
  var recipientId int
  var messageType string
  var content string
  var width sql.NullInt64
  var height sql.NullInt64
  var length sql.NullInt64
  var source sql.NullString
  var rows *sql.Rows
  if params.usePagination {
    start := params.pageToLoad * params.messagesPerPage
    end := (params.pageToLoad + 1) * params.messagesPerPage
    log.Printf("start %d end %d", start, end)
    rows, err = client.db.Query(SELECT_MESSAGES_BETWEEN_USERS_WITH_LIMIT, requestedSenderId,
                               requestedRecipientId, requestedRecipientId,
                               requestedSenderId, start, end)
  } else {
    rows, err = client.db.Query(SELECT_MESSAGES_BETWEEN_USERS, requestedSenderId,
                               requestedRecipientId, requestedRecipientId,
                               requestedSenderId)
  }
  if err != nil {
    return nil, errors.New("bad messagesPerPage or pageToLoad, no results found for desired page")
  }
  for rows.Next() {
    if err := rows.Scan(&senderId, &recipientId, &messageType, &content,
                        &width, &height, &length, &source); err != nil {
      return nil, err
    }
    sender := params.senderName
    recipient := params.recipientName
    if senderId != int(requestedSenderId) {
      sender = params.recipientName
      recipient = params.senderName
    }
    // If there is associated metadata, save it in the MessageMetadata struct.
    var metadata *MessageMetadata
    switch messageType {
    case MESSAGE_TYPE_PLAINTEXT:
      metadata = nil
      break
    case MESSAGE_TYPE_IMAGE_LINK:
      metadata = &MessageMetadata {
        Width: int(width.Int64),
        Height: int(height.Int64),
      }
      break
    case MESSAGE_TYPE_VIDEO_LINK:
      metadata = &MessageMetadata {
        Length: int(length.Int64),
        Source: source.String,
      }
      break
    default:
      // Should never get here.
      return nil, errors.New(fmt.Sprintf("Unknown message type %s", messageType))
    }
    messages = append(messages, &Message {
      Sender: sender,
      Recipient: recipient,
      MessageType: messageType,
      Content: content,
      Metadata: metadata,
    })
  }
  return messages, nil
}

// Factory for creating a new client with the given connection information.
func NewChatSqlClient(driverName string, dataSourceName string) (*ChatSQLClient, error) {
  db, err := sql.Open(driverName, dataSourceName)
  if err != nil {
    return nil, err
  }
  client := &ChatSQLClient{
    db: db,
  }
  return client, nil
}
