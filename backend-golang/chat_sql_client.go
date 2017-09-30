package main

import (
  "database/sql"
  _ "log"

  _ "github.com/go-sql-driver/mysql"
)

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
}

func (client *ChatSQLClient) GetTest() (string, error) {
  var result string
  err := client.db.QueryRow(`SELECT col FROM test`).Scan(&result)
  return result, err
}

// Factory for creating a new client with the given connection information.
func NewChatSqlClient(driverName string, dataSourceName string) (*ChatSQLClient, error) {
  db, err := sql.Open(driverName, dataSourceName)
  client := &ChatSQLClient{db: db}
  return client, err
}
