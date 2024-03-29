package main

type Stream struct {
	StreamID    uint   `json:"stream_id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Topic struct {
	MaxID uint   `json:"max_id"`
	Name  string `json:"name" gorm:"unique"`
}

type Message struct {
	MessageID      uint   `json:"id" gorm:"primaryKey"`
	Timestamp      uint   `json:"timestamp"`
	Content        string `json:"content"`
	ContentType    string `json:"content_type"`
	AvatarUrl      string `json:"avatar_url"`
	Client         string `json:"client"`
	SenderEmail    string `json:"sender_email"`
	SenderFullName string `json:"sender_full_name"`
	SenderID       uint   `json:"sender_id"`
	StreamID       uint   `json:"stream_id"`
	Subject        string `json:"subject"`
}

/*
	Message fields not added:

	display_recipient: Data on the recipient of the message; either the name of a stream or a dictionary containing data on the users who received the message.
	flags: The user's message flags for the message.
	reactions: Data on any reactions to the message.
	recipient_id: A unique ID for the set of users receiving the message (either a stream or group of users). Useful primarily for hashing.
	sender_id: The user ID of the message's sender.
	sender_realm_str: A string identifier for the realm the sender is in.
	sender_short_name: Reserved for future use.
	subject_links: Data on any links to be included in the topic line (these are generated by custom linkification filters that match content in the message's topic.)
*/

type File struct {
	Path        string `gorm:"uniqueIndex"`
	ContentType string
	Size        int
	Data        []byte
}
