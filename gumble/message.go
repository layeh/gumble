package gumble

// Message is data that be encoded and sent to the server. The following
// types implement this interface: AudioBuffer, AccessTokens, BanList,
// RegisteredUsers, TextMessage, and VoiceTarget.
type Message interface {
	writeMessage(client *Client) error
}
