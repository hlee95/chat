USE challenge;

# There are 3 tables to keep track of the data for this chat app.
# - users
# - messages
# - messages_metadata
# Each is defined and described in this file.

CREATE TABLE test(col VARCHAR(10));
INSERT INTO test(col) VALUES('ok');

# Store users and their hashed passwords and the salt used to hash.
# Usernames are limited to 10 chars.
CREATE TABLE users(
  id INT NOT NULL AUTO_INCREMENT,
  username VARCHAR(10) NOT NULL,
  hash BINARY(60) NOT NULL,
  salt BINARY(16) NOT NULL,
  PRIMARY KEY (id)
);
# Create index for username since that will be the most used query.
CREATE INDEX user_idx on users(username);

# Stores all messages.
# Message content for now is limited to 255 chars.
# Store user ids not usernames because we may want to allow changes to usernames.
CREATE TABLE messages(
  id INT NOT NULL AUTO_INCREMENT,
  sender_id INT NOT NULL,
  recipient_id INT NOT NULL,
  message_type ENUM('plaintext', 'image_link', 'video_link') NOT NULL,
  message_content TINYTEXT NOT NULL,
  message_metadata_id INT,
  PRIMARY KEY (id)
);
# Create index for sender and recipient to improve performance of recovering
# message history between two people.
CREATE INDEX sender_recipient_idx on messages(sender_id, recipient_id);

# Stores optional metadata for messages, so that not every row in the messages
# table needs to have these fields available.
# Messages that are image links need a width and height.
# Messages that are video links need a length and source.
CREATE TABLE messages_metadata (
  id INT NOT NULL AUTO_INCREMENT,
  width SMALLINT,
  height SMALLINT,
  length SMALLINT,
  source VARCHAR(16),
  PRIMARY KEY (id)
);
