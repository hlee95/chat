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

const SELECT_ID_FROM_USERNAME = "SELECT id FROM users WHERE username=?"
const SELECT_USERNAME_FROM_ID = "SELECT username FROM users WHERE id=?"
const SELECT_IMAGE_METADATA = "SELECT width, height FROM messages_metadata WHERE id=?"
const SELECT_VIDEO_METADATA = "SELECT length, source FROM messages_metadata WHERE id=?"
const SELECT_MESSAGES_BETWEEN_USERS = "SELECT sender_id, recipient_id, message_type, message_content, message_metadata_id FROM messages WHERE (sender_id=? AND recipient_id=?) OR (sender_id=? AND recipient_id=?) ORDER BY id"
const SELECT_USER_CREDENTIALS = "SELECT hash, salt FROM users WHERE username=?"

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

// Given a user, get its id.
func (client *ChatSQLClient) getUserId(username string) (int64, error) {
  var id int64
  err := client.db.QueryRow(SELECT_ID_FROM_USERNAME, username).Scan(&id)
  return id, err
}

// Get a username by id.
func (client *ChatSQLClient) getUserById(id int) (string, error) {
  var username string
  err := client.db.QueryRow(SELECT_USERNAME_FROM_ID, id).Scan(&username)
  return username, err
}

// Get image metadata from its database id.
func (client *ChatSQLClient) getImageMetadataFromId(id int) (*MessageMetadata, error) {
  var width int
  var height int
  if err := client.db.QueryRow(SELECT_IMAGE_METADATA, id).Scan(&width, &height); err != nil {
    return nil, err
  }
  metadata := &MessageMetadata{
    Width: width,
    Height: height,
  }
  return metadata, nil
}

// Get video metadata from its database id.
func (client *ChatSQLClient) getVideoMetadataFromId(id int) (*MessageMetadata, error) {
  var length int
  var source string
  if err := client.db.QueryRow(SELECT_VIDEO_METADATA, id).Scan(&length, &source); err != nil {
    return nil, err
  }
  metadata := &MessageMetadata{
    Length: length,
    Source: source,
  }
  return metadata, nil
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

// Returns true if the given username already exists in the database.
func (client *ChatSQLClient) CheckUserExists(username string) bool {
  _, err := client.getUserId(username)
  return err == nil
}

// Retrieves the password hash and salt for the given username.
func (client *ChatSQLClient) GetUserCredentials(username string) (hash_ []byte, salt_ []byte, err_ error) {
  var hash []byte
  var salt []byte
  err := client.db.QueryRow(SELECT_USER_CREDENTIALS, username).Scan(&hash, &salt)
  if err != nil {
    return nil, nil, err
  }
  return hash, salt, nil
}

// Adds a new message to the database. Returns the id of that message, or an error.
func (client *ChatSQLClient) AddMessage(senderName string, recipientName string, messageType string, content string) (id int64, err_ error) {
  // Find the associated ids of the two users.
  var err error
  senderId, err := client.getUserId(senderName)
  if err != nil {
    return -1, err
  }
  recipientId, err := client.getUserId(recipientName)
  if err != nil {
    return -1, err
  }
  switch messageType {
  case MESSAGE_TYPE_PLAINTEXT:
    res, err := client.db.Exec(INSERT_MESSAGE_WITH_NO_METADATA, senderId, recipientId, messageType, content)
    if err != nil {
      return -1, err
    }
    return res.LastInsertId()
  case MESSAGE_TYPE_IMAGE_LINK:
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
  case MESSAGE_TYPE_VIDEO_LINK:
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
    return -1, errors.New(fmt.Sprintf("Unknown message type %s", messageType))
  }
}

// Gets messages between two users.
func (client *ChatSQLClient) FetchMessages(senderName string, recipientName string) ([]*Message, error) {
  // Find the associated ids of the two users.
  var err error
  requestedSenderId, err := client.getUserId(senderName)
  if err != nil {
    return make([]*Message, 0), err
  }
  requestedRecipientId, err := client.getUserId(recipientName)
  if err != nil {
    return make([]*Message, 0), err
  }
  // Get all rows.
  var messages []*Message
  var senderId int
  var recipientId int
  var messageType string
  var content string
  var metadata_id sql.NullInt64
  rows, err := client.db.Query(SELECT_MESSAGES_BETWEEN_USERS, requestedSenderId, requestedRecipientId, requestedRecipientId, requestedSenderId)
  if err != nil {
    return make([]*Message, 0), err
  }
  for rows.Next() {
    if err := rows.Scan(&senderId, &recipientId, &messageType, &content, &metadata_id); err != nil {
      return make([]*Message, 0), err
    }
    sender, err := client.getUserById(senderId)
    if err != nil {
      return make([]*Message, 0), err
    }
    recipient, err := client.getUserById(recipientId)
    if err != nil {
      return make([]*Message, 0), err
    }
    // If there is associated metadata, fetch it.
    if metadata_id.Valid {
      switch messageType {
      case MESSAGE_TYPE_IMAGE_LINK:
        metadata, err := client.getImageMetadataFromId(int(metadata_id.Int64))
        if err != nil {
          return make([]*Message, 0), err
        }
        messages = append(messages, &Message {
          Sender: sender,
          Recipient: recipient,
          MessageType: messageType,
          Content: content,
          Metadata: metadata,
        })
      case MESSAGE_TYPE_VIDEO_LINK:
        metadata, err := client.getVideoMetadataFromId(int(metadata_id.Int64))
        if err != nil {
          return make([]*Message, 0), err
        }
        messages = append(messages, &Message {
          Sender: sender,
          Recipient: recipient,
          MessageType: messageType,
          Content: content,
          Metadata: metadata,
        })
      default:
        // Should never get here.
        return make([]*Message, 0), errors.New(fmt.Sprintf("Unknown message type %s", messageType))
      }
    } else {
      // No metadata, so just add in the normal message fields.
      messages = append(messages, &Message {
        Sender: sender,
        Recipient: recipient,
        MessageType: messageType,
        Content: content,
        Metadata: nil,
      })
    }
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
