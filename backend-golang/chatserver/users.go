package chatserver

import (
  "encoding/json"
  "log"
  "net/http"
  "strconv"

  auth "app/chatauth"
)

type createUserStruct struct {
  Username string;
  Password string;
}

// Request handler for /users.
// Expects a POST request with the following parameters in the body
// - Username : maximum 10 characters
// - Password : maximum 72 characters (due to bcrypt limitation)
// Expects data in JSON, because it's easier to send JSON than url-encoded
// key value pairs in React, and our frontend is in React.
//
// Sample curl request:
// curl -d '{"Username":"hanna", "Password":"secret"}' -H "Content-Type: application/json" -X POST localhost:18000/users
func (server *ChatServer) handleUsers(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "application/json")
  switch r.Method {
  case http.MethodPost:
    // Parse request.
    var body createUserStruct
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&body); err != nil {
      log.Printf("Bad POST request received at /users, %+v", r)
      http.Error(w, "bad POST request, could not parse", http.StatusBadRequest)
      return
    }
    username := body.Username
    password := body.Password
    log.Printf("Received POST at /users for user %s", username)
    // Check lengths of username and password.
    if len(username) > 10 {
      log.Printf("Username too long")
      http.Error(w, "max username length is 10 characters", http.StatusUnprocessableEntity)
      return
    }
    if len(password) > 72 {
      log.Printf("Password too long")
      http.Error(w, "max password length is 72 characters", http.StatusUnprocessableEntity)
      return
    }
    // Respond with StatusConflict (409) if username already exists.
    if server.db.CheckUserExists(username) {
      log.Printf("User %s already exists", username)
      http.Error(w, "username already exists (case insensitive)", http.StatusConflict)
      return
    }
    // Hash password and create a new user.
    salt := auth.GenerateSalt(16)
    hash, err := auth.HashPasswordWithSalt(password, salt)
    if err != nil {
      log.Printf("Error hashing password, %s", err.Error())
      http.Error(w, "hashing error", http.StatusInternalServerError)
      return
    }
    id, err := server.db.CreateUser(username, hash, salt)
    if err != nil {
      log.Printf("Error creating a user, %s", err.Error())
      http.Error(w, "database error", http.StatusInternalServerError)
      return
    }
    // Success!
    log.Printf("User %s created successfully, id %d", username, id)
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(map[string]string{
      "username": username,
      "id": strconv.FormatInt(id, 10),
    }); err != nil {
      log.Printf("Error formatting http response, %s", err.Error())
      http.Error(w, "error generating response", http.StatusInternalServerError)
    }
  default:
    // Unhandled request, respond with StatusMethodNotAllowed (405).
    log.Printf("Unknown request received at /users, %+v", r)
    http.Error(w, "only POST requests are accepted", http.StatusMethodNotAllowed)
  }
}
