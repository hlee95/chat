package chatserver

import (
  "log"
  "net/http"
)

const DRIVER_NAME = "mysql"
const DATA_SOURCE_NAME = "root:testpass@tcp(db:3306)/challenge"

// ChatServer maintains a db connection and any relevant state,
// and responds to HTTP requests.
type ChatServer struct {
  db *ChatSQLClient
}

// Startup. Should be called by main.
func (server *ChatServer) Start() {
  // Make db connection.
  db, err := NewChatSqlClient(DRIVER_NAME, DATA_SOURCE_NAME)
  if err != nil {
    log.Fatal("unable to connect to DB: ", err)
  }
  server.db = db

  // Assign handlers for requests we accept.
  http.HandleFunc("/users", server.handleUsers)
  http.HandleFunc("/messages", server.handleMessages)
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
  })

  // Begin serving, fail on any errors.
  if err := http.ListenAndServe(":8000", nil); err != nil {
    log.Fatal(err)
  }
}

