package chatserver

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

const HASH_SIZE = 60
const SALT_SIZE = 16
const PREPARE_ADD_USER = "INSERT INTO users(username, hash, salt) VALUES(?, ?, ?)"

// ChatSQLClient wraps a connection to the database, and provides an
// api to the server.
// This abstraction is in case we want to swap out a different db
// for the project without affecting the logic in the server.
//
// API exposed to server includes the following:
// - GetHashedPasswordAndSalt(username string)
// - GetMessages(sender, receiver)
// - PutMessage(sender, receiver, messageType, messageContent, properties)
//
// ** Note that the server is responsible for handling errors propagated
// up by the db client. **
type ChatSQLClient struct {
  db *sql.DB
  prepare_add_user *sql.Stmt
}

func (client *ChatSQLClient) Test() (string, error) {
  var result string
  err := client.db.QueryRow(`SELECT col FROM test`).Scan(&result)
  return result, err
}

// Create a new user in the database with the given username, password hash, and salt.
// Returns the id of the newly created user, or an error.
func (client *ChatSQLClient) CreateUser(username string, hash []byte, salt []byte) (id int64, err error) {
  tx, err := client.db.Begin()
  if err != nil {
    return -1, err
  }
  res, err := tx.Stmt(client.prepare_add_user).Exec(username, hash, salt)
  if err != nil {
    return -1, err
  }
  err = tx.Commit()
  if err != nil {
    return -1, err
  }
  return res.LastInsertId()
}

// Returns true if the given username already exists in the database.
func (client *ChatSQLClient) CheckUserExists(username string) bool {
  var id int
  err := client.db.QueryRow(`SELECT id FROM users WHERE username=?`, username).Scan(&id)
  // If there is an error, then no row was found, so return false.
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

// Factory for creating a new client with the given connection information.
func NewChatSqlClient(driverName string, dataSourceName string) (*ChatSQLClient, error) {
  db, err := sql.Open(driverName, dataSourceName)
  // Set up prepared statements.
  prepare_add_user, err := db.Prepare(PREPARE_ADD_USER)
  if err != nil {
    return nil, err
  }
  client := &ChatSQLClient{
    db: db,
    prepare_add_user: prepare_add_user}
  return client, err
}
