package chatserver

// This file defines common structs and constants used in the chatserver package.
// Make sure field names are capitalized, otherwise JSON encoder won't work.

// Message types.
const MESSAGE_TYPE_PLAINTEXT = "plaintext"
const MESSAGE_TYPE_IMAGE_LINK = "image_link"
const MESSAGE_TYPE_VIDEO_LINK = "video_link"

// Defines a message.
type Message struct {
  Sender      string;
  Recipient   string;
  MessageType string;
  Content     string;
  Metadata    *MessageMetadata;
}

// Defines message metadata.
type MessageMetadata struct {
  Width       int;
  Height      int;
  Length      int;
  Source      string;
}

// Hardcode message metadata for images and videos for now for simplicity.
const IMAGE_WIDTH = 100
const IMAGE_HEIGHT = 200
const VIDEO_LENGTH = 300
const VIDEO_SOURCE = "YouTube"
