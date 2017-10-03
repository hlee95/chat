package chatserver

// This file defines common structs and constants used in the chatserver package.
// Field names must be capitalized, otherwise JSON encoder won't work.
// However, we can provide lowercase identifiers so that clients don't need
// to deal with the capitalization.

// Message types.
const MESSAGE_TYPE_PLAINTEXT = "plaintext"
const MESSAGE_TYPE_IMAGE_LINK = "image_link"
const MESSAGE_TYPE_VIDEO_LINK = "video_link"

// Defines a message.
type Message struct {
  Sender      string           `json:"sender"`
  Recipient   string           `json:"recipient"`
  MessageType string           `json:"messageType"`
  Content     string           `json:"content"`
  Metadata    *MessageMetadata `json:"metadata"`
}

// Defines message metadata.
type MessageMetadata struct {
  Width       int    `json:"width"`
  Height      int    `json:"height"`
  Length      int    `json:"length"`
  Source      string `json:"source"`
}

// Hardcode message metadata for images and videos for now for simplicity.
const IMAGE_WIDTH = 100
const IMAGE_HEIGHT = 200
const VIDEO_LENGTH = 300
const VIDEO_SOURCE = "YouTube"
