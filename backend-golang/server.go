package main

import (
  "encoding/json"
  "log"
  "net/http"
  _ "golang.org/x/crypto/bcrypt"
)

const DRIVER_NAME = "mysql"
const DATA_SOURCE_NAME = "root:testpass@tcp(db:3306)/challenge"

// ChatServer maintains a db connection and any relevant state,
// and responds to HTTP requests. ChatServer is also responsible
// for encrypting, decrypting or hashing data as is necessary.
type ChatServer struct {
  db *ChatSQLClient
}

// Startup. Should be called by main.
func (server *ChatServer) Start() {
  // Make db connection.
  db, err := NewChatSqlClient(DRIVER_NAME, DATA_SOURCE_NAME)
  if err != nil {
    log.Fatal("unable to connect to DB", err)
  }
  server.db = db

  // Assign handlers for requests we accept.
  http.HandleFunc("/test", server.handleTest)
  http.HandleFunc("/users", server.handleUsers)
  http.HandleFunc("/messages", server.handleMessages)

  // Begin serving, fail on any errors.
  if err := http.ListenAndServe(":8000", nil); err != nil {
    log.Fatal(err)
  }
}

// Request handler for /test.
func (server *ChatServer) handleTest(w http.ResponseWriter, r *http.Request) {
  result, err := server.db.GetTest()
  if err != nil {
    log.Fatal(err)
  }

  w.Header().Add("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(map[string]string{
    "result":  result,
    "backend": "go",
  }); err != nil {
    log.Panic(err)
  }
}

// Request handler for /users.
func (server *ChatServer) handleUsers(w http.ResponseWriter, r *http.Request) {
  // TODO
}

// Request handler for /messages
func (server *ChatServer) handleMessages(w http.ResponseWriter,
                                         r *http.Request) {
  // TODO
}

