package gumble

import (
	"errors"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/proto"
)

type handlerFunc func(*Client, []byte) error

var (
	errUnimplementedHandler = errors.New("the handler has not been implemented")
	errIncompleteProtobuf   = errors.New("protobuf message is missing a required field")
	errInvalidProtobuf      = errors.New("protobuf message has an invalid field")
)

var handlers map[uint16]handlerFunc

func init() {
	handlers = map[uint16]handlerFunc{
		0:  handleVersion,
		1:  handleUdpTunnel,
		2:  handleAuthenticate,
		3:  handlePing,
		4:  handleReject,
		5:  handleServerSync,
		6:  handleChannelRemove,
		7:  handleChannelState,
		8:  handleUserRemove,
		9:  handleUserState,
		10: handleBanList,
		11: handleTextMessage,
		12: handlePermissionDenied,
		13: handleAcl,
		14: handleQueryUsers,
		15: handleCryptSetup,
		16: handleContextActionModify,
		17: handleContextAction,
		18: handleUserList,
		19: handleVoiceTarget,
		20: handlePermissionQuery,
		21: handleCodecVersion,
		22: handleUserStats,
		23: handleRequestBlob,
		24: handleServerConfig,
		25: handleSuggestConfig,
	}
}

func handleVersion(client *Client, buffer []byte) error {
	var packet MumbleProto.Version
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	var version Version
	if packet.Version != nil {
		version.version = *packet.Version
	}
	if packet.Release != nil {
		version.release = *packet.Release
	}
	if packet.Os != nil {
		version.os = *packet.Os
	}
	if packet.OsVersion != nil {
		version.osVersion = *packet.OsVersion
	}
	client.server.version = version

	return nil
}

func handleUdpTunnel(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleAuthenticate(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handlePing(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleReject(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleServerSync(client *Client, buffer []byte) error {
	var packet MumbleProto.ServerSync
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := &ConnectEvent{}

	if packet.Session != nil {
		client.self = client.users.BySession(uint(*packet.Session))
	}
	if packet.WelcomeText != nil {
		event.WelcomeMessage = *packet.WelcomeText
	}
	// TODO: bandwidth, permissions
	client.state = Synced

	client.listeners.OnConnect(event)
	return nil
}

func handleChannelRemove(client *Client, buffer []byte) error {
	var packet MumbleProto.ChannelRemove
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.ChannelId == nil {
		return errIncompleteProtobuf
	}
	var channel *Channel
	{
		channelId := uint(*packet.ChannelId)
		channel = client.channels.ById(channelId)
		if channel == nil {
			return errInvalidProtobuf
		}
		client.channels.Delete(channelId)
	}

	if client.state == Synced {
		event := &ChannelChangeEvent{
			Channel: channel,
		}
		client.listeners.OnChannelChange(event)
	}
	return nil
}

func handleChannelState(client *Client, buffer []byte) error {
	var packet MumbleProto.ChannelState
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.ChannelId == nil {
		return errIncompleteProtobuf
	}
	var channel *Channel
	{
		channelId := uint(*packet.ChannelId)
		if !client.channels.Exists(channelId) {
			channel = client.channels.Create(channelId)
			channel.client = client
		} else {
			channel = client.channels.ById(channelId)
		}
	}
	if packet.Parent != nil {
		channel.parent = client.channels.ById(uint(*packet.Parent))
	}
	if packet.Name != nil {
		channel.name = *packet.Name
	}
	if packet.Description != nil {
		channel.description = *packet.Description
		channel.descriptionHash = nil
	}
	if packet.Temporary != nil {
		channel.temporary = *packet.Temporary
	}
	if packet.Position != nil {
		channel.position = *packet.Position
	}
	if packet.DescriptionHash != nil {
		channel.descriptionHash = packet.DescriptionHash
		channel.description = ""
	}
	// TODO: channel links

	if client.state == Synced {
		event := &ChannelChangeEvent{
			Channel: channel,
		}
		client.listeners.OnChannelChange(event)
	}
	return nil
}

func handleUserRemove(client *Client, buffer []byte) error {
	var packet MumbleProto.UserRemove
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	var user *User
	{
		session := uint(*packet.Session)
		user = client.users.BySession(session)
		if user == nil {
			return errInvalidProtobuf
		}
		client.users.Delete(session)
	}

	if client.state == Synced {
		event := &UserChangeEvent{
			User: user,
		}
		client.listeners.OnUserChange(event)
	}
	return nil
}

func handleUserState(client *Client, buffer []byte) error {
	var packet MumbleProto.UserState
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	var user, actor *User
	{
		session := uint(*packet.Session)
		if !client.users.Exists(session) {
			user = client.users.Create(session)
			user.channel = client.channels.ById(0)
			user.client = client
		} else {
			user = client.users.BySession(session)
		}
	}
	if packet.Actor != nil {
		actor = client.users.BySession(uint(*packet.Actor))
		if actor == nil {
			return errInvalidProtobuf
		}
	}
	if packet.Name != nil {
		user.name = *packet.Name
	}
	if packet.UserId != nil {
		user.userId = *packet.UserId
	}
	if packet.ChannelId != nil {
		user.channel = client.channels.ById(uint(*packet.ChannelId))
	}
	if packet.Mute != nil {
		user.mute = *packet.Mute
	}
	if packet.Deaf != nil {
		user.deaf = *packet.Deaf
	}
	if packet.Suppress != nil {
		user.suppress = *packet.Suppress
	}
	if packet.SelfMute != nil {
		user.selfMute = *packet.SelfMute
	}
	if packet.SelfDeaf != nil {
		user.selfDeaf = *packet.SelfDeaf
	}
	if packet.Texture != nil {
		user.texture = packet.Texture
		user.textureHash = nil
	}
	if packet.Comment != nil {
		user.comment = *packet.Comment
		user.commentHash = nil
	}
	if packet.Hash != nil {
		user.hash = *packet.Hash
	}
	if packet.CommentHash != nil {
		user.commentHash = packet.CommentHash
		user.comment = ""
	}
	if packet.TextureHash != nil {
		user.textureHash = packet.TextureHash
		user.texture = nil
	}
	if packet.PrioritySpeaker != nil {
		user.prioritySpeaker = *packet.PrioritySpeaker
	}
	if packet.Recording != nil {
		user.recording = *packet.Recording
	}

	if client.state == Synced {
		event := &UserChangeEvent{
			User: user,
		}
		client.listeners.OnUserChange(event)
	}
	return nil
}

func handleBanList(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleTextMessage(client *Client, buffer []byte) error {
	var packet MumbleProto.TextMessage
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := &TextMessageEvent{}
	if packet.Actor != nil {
		event.Sender = client.users.BySession(uint(*packet.Actor))
		// TODO: ensure non-nil
	}
	if packet.Session != nil {
		event.Users = make([]*User, len(packet.Session))
		for i, session := range packet.Session {
			event.Users[i] = client.users.BySession(uint(session))
			// TODO: ensure non-nil
		}
	}
	if packet.ChannelId != nil {
		event.Channels = make([]*Channel, len(packet.ChannelId))
		for i, id := range packet.ChannelId {
			event.Channels[i] = client.channels.ById(uint(id))
			// TODO: ensure non-nil
		}
	}
	if packet.Message != nil {
		event.Message = *packet.Message
	}
	// TODO: trees

	client.listeners.OnTextMessage(event)
	return nil
}

func handlePermissionDenied(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleAcl(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleQueryUsers(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleCryptSetup(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleContextActionModify(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleContextAction(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleUserList(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleVoiceTarget(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handlePermissionQuery(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleCodecVersion(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleUserStats(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleRequestBlob(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleServerConfig(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}

func handleSuggestConfig(client *Client, buffer []byte) error {
	// TODO
	return errUnimplementedHandler
}
