package chatserver

import (
  "database/sql"
  "errors"
  "fmt"
  _ "github.com/go-sql-driver/mysql"
)

const INSERT_USER = "INSERT INTO users(username, hash, salt) VALUES(?, ?, ?)"
const INSERT_MESSAGE = "INSERT INTO messages(sender_id, recipient_id, message_type, message_content, message_metadata_id) VALUES (?, ?, ?, ?, ?)"
const INSERT_MESSAGE_WITH_NO_METADATA = "INSERT INTO messages(sender_id, recipient_id, message_type, message_content) VALUES (?, ?, ?, ?)"
const INSERT_MESSAGES_IMAGE_METADATA = "INSERT INTO messages_metadata(width, height) VALUES(?, ?)"
const INSERT_MESSAGES_VIDEO_METADATA = "INSERT INTO messages_metadata(length, source) VALUES(?, ?)"

// Hardcode message metadata for images and videos for now for simplicity.
const IMAGE_WIDTH = 100
const IMAGE_HEIGHT = 200
const VIDEO_LENGTH = 300
const VIDEO_SOURCE = "YouTube"

// ChatSQLClient wraps a connection to the database, and provides an
// api to the server.
// This abstraction is in case we want to swap out a different db
// for the project without affecting the logic in the server.
//
// API exposed to server includes the following:
// - NewChatSqlClient()
// - client.CreateUser(username)
// - client.GetUserId(username)
// - client.CheckUserExists(username)
// - client.GetUserCredentials(username)
// - client.GetMessages(sender, recipient)
// - client.AddMessage(senderName, recipientName, messageType, messageContent, properties)
//
// ** Note that the server is responsible for handling errors propagated
// up by the db client. **
type ChatSQLClient struct {
  db *sql.DB
}

func (client *ChatSQLClient) Test() (string, error) {
  var result string
  err := client.db.QueryRow(`SELECT col FROM test`).Scan(&result)
  return result, err
}

// Create a new user in the database with the given username, password hash, and salt.
// Returns the id of the newly created user, or an error.
func (client *ChatSQLClient) CreateUser(username string, hash []byte, salt []byte) (id int64, err error) {
  res, err := client.db.Exec(INSERT_USER, username, hash, salt)
  if err != nil {
    return -1, err
  }
  return res.LastInsertId()
}

// Given a user, get its id.
func (client *ChatSQLClient) GetUserId(username string) (int64, error) {
  var id int64
  err := client.db.QueryRow(`SELECT id FROM users WHERE username=?`, username).Scan(&id)
  return id, err
}

// Returns true if the given username already exists in the database.
func (client *ChatSQLClient) CheckUserExists(username string) bool {
  _, err := client.GetUserId(username)
  return err == nil
}

// Retrieves the password hash and salt for the given username.
func (client *ChatSQLClient) GetUserCredentials(username string) (hash_ []byte, salt_ []byte, err_ error) {
  var hash []byte
  var salt []byte
  err := client.db.QueryRow(`SELECT hash, salt FROM users WHERE username=?`, username).Scan(&hash, &salt)
  if err != nil {
    return nil, nil, err
  }
  return hash, salt, nil
}

// Adds a new message to the database. Returns the id of that message, or an error.
func (client *ChatSQLClient) AddMessage(senderName string, recipientName string, messageType string, content string) (id int64, err_ error) {
  // Find the associated ids of the two users.
  var err error
  senderId, err := client.GetUserId(senderName)
  if err != nil {
    return -1, err
  }
  recipientId, err := client.GetUserId(recipientName)
  if err != nil {
    return -1, err
  }
  switch messageType {
  case "plaintext":
    res, err := client.db.Exec(INSERT_MESSAGE_WITH_NO_METADATA, senderId, recipientId, messageType, content)
    if err != nil {
      return -1, err
    }
    return res.LastInsertId()
  case "image_link":
    // First insert the metadata.
    res, err := client.db.Exec(INSERT_MESSAGES_IMAGE_METADATA, IMAGE_WIDTH, IMAGE_HEIGHT)
    metadataId, err := res.LastInsertId()
    if err != nil {
      return -1, err
    }
    // Insert the message.
    res, err = client.db.Exec(INSERT_MESSAGE, senderId, recipientId, messageType, content, metadataId)
    if err != nil {
      return -1, err
    }
    return res.LastInsertId()
  case "video_link":
    // First insert the metadata.
    res, err := client.db.Exec(INSERT_MESSAGES_VIDEO_METADATA, VIDEO_LENGTH, VIDEO_SOURCE)
    metadataId, err := res.LastInsertId()
    if err != nil {
      return -1, err
    }
    // Insert the message.
    res, err = client.db.Exec(INSERT_MESSAGE, senderId, recipientId, messageType, content, metadataId)
    if err != nil {
      return -1, err
    }
    return res.LastInsertId()
  default:
    return -1, errors.New(fmt.Sprintf("unknown message type %s", messageType))
  }
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
