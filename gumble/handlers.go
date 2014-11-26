package gumble

import (
	"bytes"
	"errors"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gopus"
	"github.com/bontibon/gumble/gumble/MumbleProto"
	"github.com/bontibon/gumble/gumble/varint"
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

func parseVersion(packet *MumbleProto.Version) Version {
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
	return version
}

func handleVersion(client *Client, buffer []byte) error {
	var packet MumbleProto.Version
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	client.server.version = parseVersion(&packet)
	return nil
}

func handleUdpTunnel(client *Client, buffer []byte) error {
	if client.audio == nil || !client.audio.IsSink() {
		return nil
	}

	reader := bytes.NewReader(buffer)
	var bytesRead int64

	var audioType byte
	var audioTarget byte
	var sequence int64
	var user *User
	var audioLength int

	// Header byte
	if typeTarget, err := varint.ReadByte(reader); err != nil {
		return err
	} else {
		audioType = (typeTarget >> 5) & 0x7
		audioTarget = typeTarget & 0x1F
		// Opus only
		if audioType != 4 {
			return errInvalidProtobuf
		}
		bytesRead += 1
	}

	// Session
	if session, n, err := varint.ReadFrom(reader); err != nil {
		return err
	} else {
		user = client.users[uint(session)]
		if user == nil {
			return errInvalidProtobuf
		}
		bytesRead += n
	}

	// Sequence
	if seq, n, err := varint.ReadFrom(reader); err != nil {
		return err
	} else {
		sequence = seq
		bytesRead += n
	}

	// Length
	if length, n, err := varint.ReadFrom(reader); err != nil {
		return err
	} else {
		audioLength = int(length)
		if audioLength > reader.Len() {
			return errInvalidProtobuf
		}
		bytesRead += n
	}

	opus := buffer[bytesRead : bytesRead+int64(audioLength)]
	if pcm, err := user.decoder.Decode(opus, MaximumFrameSize, false); err != nil {
		return err
	} else {
		_ = audioTarget
		packet := AudioPacket{
			Sender:   user,
			Sequence: int(sequence),
			Pcm:      pcm,
		}
		client.audio.incoming(&packet)
	}
	return nil
}

func handleAuthenticate(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handlePing(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleReject(client *Client, buffer []byte) error {
	var packet MumbleProto.Reject
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := DisconnectEvent{
		Client: client,
	}
	if packet.Type != nil {
		event.Type = DisconnectType(*packet.Type)
	}
	if packet.Reason != nil {
		event.String = *packet.Reason
	}
	client.close(&event)
	return nil
}

func handleServerSync(client *Client, buffer []byte) error {
	var packet MumbleProto.ServerSync
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := ConnectEvent{
		Client: client,
	}

	if packet.Session != nil {
		client.self = client.users.BySession(uint(*packet.Session))
	}
	if packet.WelcomeText != nil {
		event.WelcomeMessage = *packet.WelcomeText
	}
	if packet.MaxBandwidth != nil {
		event.MaximumBitrate = int(*packet.MaxBandwidth)
	}
	client.state = StateSynced

	client.listeners.OnConnect(&event)
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
		if parent := channel.parent; parent != nil {
			channel.parent.children.Delete(uint(channel.id))
		}
	}

	if client.state == StateSynced {
		event := ChannelChangeEvent{
			Client:  client,
			Channel: channel,
			Removed: true,
		}
		client.listeners.OnChannelChange(&event)
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
	event := ChannelChangeEvent{
		Client: client,
	}
	var channel *Channel
	channelId := uint(*packet.ChannelId)
	if !client.channels.Exists(channelId) {
		channel = client.channels.Create(channelId)
		channel.client = client

		event.Created = true
	} else {
		channel = client.channels.ById(channelId)
	}
	event.Channel = channel
	if packet.Parent != nil {
		if channel.parent != nil {
			channel.parent.children.Delete(channelId)
		}
		newParent := client.channels.ById(uint(*packet.Parent))
		if newParent != channel.parent {
			event.Moved = true
		}
		channel.parent = newParent
		if channel.parent != nil {
			channel.parent.children[uint(channel.id)] = channel
		}
	}
	if packet.Name != nil {
		newName := *packet.Name
		if newName != channel.name {
			event.NameChanged = true
		}
		channel.name = newName
	}
	if packet.Description != nil {
		newDescription := *packet.Description
		if newDescription != channel.description {
			event.DescriptionChanged = true
		}
		channel.description = newDescription
		channel.descriptionHash = nil
	}
	if packet.Temporary != nil {
		channel.temporary = *packet.Temporary
	}
	if packet.Position != nil {
		channel.position = *packet.Position
	}
	if packet.DescriptionHash != nil {
		event.DescriptionChanged = true
		channel.descriptionHash = packet.DescriptionHash
		channel.description = ""
	}

	if client.state == StateSynced {
		client.listeners.OnChannelChange(&event)
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
		if user.channel != nil {
			user.channel.users.Delete(session)
		}
		client.users.Delete(session)
	}

	if client.state == StateSynced {
		event := UserChangeEvent{
			Client:       client,
			User:         user,
			Disconnected: true,
		}
		client.listeners.OnUserChange(&event)
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
	event := UserChangeEvent{
		Client: client,
	}
	var user, actor *User
	{
		session := uint(*packet.Session)
		if !client.users.Exists(session) {
			user = client.users.Create(session)
			user.channel = client.channels.ById(0)
			user.client = client

			decoder, _ := gopus.NewDecoder(SampleRate, 1)
			user.decoder = decoder

			if user.channel == nil {
				return errInvalidProtobuf
			}
			event.ChannelChanged = true
			user.channel.users[session] = user

			event.Connected = true
		} else {
			user = client.users.BySession(session)
		}
	}
	event.User = user
	if packet.Actor != nil {
		actor = client.users.BySession(uint(*packet.Actor))
		if actor == nil {
			return errInvalidProtobuf
		}
		event.Actor = actor
	}
	if packet.Name != nil {
		newName := *packet.Name
		if newName != user.name {
			event.NameChanged = true
		}
		user.name = newName
	}
	if packet.UserId != nil {
		user.userId = *packet.UserId
	}
	if packet.ChannelId != nil {
		if user.channel != nil {
			user.channel.users.Delete(user.Session())
		}
		newChannel := client.channels.ById(uint(*packet.ChannelId))
		if newChannel == nil {
			return errInvalidProtobuf
		}
		if newChannel != user.channel {
			event.ChannelChanged = true
			user.channel = newChannel
			user.channel.users[user.Session()] = user
		}
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
		newComment := *packet.Comment
		if newComment != user.comment {
			event.CommentChanged = true
		}
		user.comment = newComment
		user.commentHash = nil
	}
	if packet.Hash != nil {
		user.hash = *packet.Hash
	}
	if packet.CommentHash != nil {
		event.CommentChanged = true
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

	if client.state == StateSynced {
		client.listeners.OnUserChange(&event)
	}
	return nil
}

func handleBanList(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleTextMessage(client *Client, buffer []byte) error {
	var packet MumbleProto.TextMessage
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := TextMessageEvent{
		Client: client,
	}
	if packet.Actor != nil {
		event.Sender = client.users.BySession(uint(*packet.Actor))
	}
	if packet.Session != nil {
		event.Users = make([]*User, 0, len(packet.Session))
		for _, session := range packet.Session {
			if user := client.users.BySession(uint(session)); user != nil {
				event.Users = append(event.Users, user)
			}
		}
	}
	if packet.ChannelId != nil {
		event.Channels = make([]*Channel, 0, len(packet.ChannelId))
		for _, id := range packet.ChannelId {
			if channel := client.channels.ById(uint(id)); channel != nil {
				event.Channels = append(event.Channels, channel)
			}
		}
	}
	if packet.TreeId != nil {
		event.Trees = make([]*Channel, 0, len(packet.TreeId))
		for _, id := range packet.TreeId {
			if channel := client.channels.ById(uint(id)); channel != nil {
				event.Trees = append(event.Trees, channel)
			}
		}
	}
	if packet.Message != nil {
		event.Message = *packet.Message
	}

	client.listeners.OnTextMessage(&event)
	return nil
}

func handlePermissionDenied(client *Client, buffer []byte) error {
	var packet MumbleProto.PermissionDenied
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Type == nil || *packet.Type == MumbleProto.PermissionDenied_H9K {
		return errInvalidProtobuf
	}

	event := PermissionDeniedEvent{
		Client: client,
		Type:   PermissionDeniedType(*packet.Type),
	}
	if packet.Reason != nil {
		event.String = *packet.Reason
	}
	if packet.Name != nil {
		event.String = *packet.Name
	}
	if packet.Session != nil {
		event.User = client.users.BySession(uint(*packet.Session))
		if event.User == nil {
			return errInvalidProtobuf
		}
	}
	if packet.ChannelId != nil {
		event.Channel = client.channels.ById(uint(*packet.ChannelId))
		if event.Channel == nil {
			return errInvalidProtobuf
		}
	}
	if packet.Permission != nil {
		event.Permission = Permission(*packet.Permission)
	}

	client.listeners.OnPermissionDenied(&event)
	return nil
}

func handleAcl(client *Client, buffer []byte) error {
	var packet MumbleProto.ACL
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := AclEvent{
		Client: client,
	}
	if packet.ChannelId != nil {
		event.Acl.channel = client.channels.ById(uint(*packet.ChannelId))
	}

	if packet.Groups != nil {
		event.Acl.groups = make([]*AclGroup, 0, len(packet.Groups))
		for _, group := range packet.Groups {
			event.Acl.groups = append(event.Acl.groups, &AclGroup{
				name: *group.Name,
			})
		}
	}

	client.listeners.OnAcl(&event)
	return nil
}

func handleQueryUsers(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleCryptSetup(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleContextActionModify(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleContextAction(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleUserList(client *Client, buffer []byte) error {
	var packet MumbleProto.UserList
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := UserListEvent{
		Client: client,
		Users:  make(RegisteredUsers, 0, len(packet.Users)),
	}

	for _, user := range packet.Users {
		registeredUser := &RegisteredUser{
			userId: *user.UserId,
		}
		if user.Name != nil {
			registeredUser.name = *user.Name
		}
		event.Users = append(event.Users, registeredUser)
	}

	client.listeners.OnUserList(&event)
	return nil
}

func handleVoiceTarget(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handlePermissionQuery(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleCodecVersion(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleUserStats(client *Client, buffer []byte) error {
	var packet MumbleProto.UserStats
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	user := client.users.BySession(uint(*packet.Session))
	if user == nil {
		return errInvalidProtobuf
	}

	if packet.Version != nil {
		user.stats.version = parseVersion(packet.Version)
	}
	if packet.Onlinesecs != nil {
		user.stats.connected = time.Now().Add(time.Duration(*packet.Onlinesecs) * -time.Second)
	}
	if packet.Idlesecs != nil {
		user.stats.idle = time.Duration(*packet.Idlesecs) * time.Second
	}

	user.statsFetched = true

	event := UserChangeEvent{
		Client:       client,
		User:         user,
		StatsChanged: true,
	}
	client.listeners.OnUserChange(&event)
	return nil
}

func handleRequestBlob(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleServerConfig(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleSuggestConfig(client *Client, buffer []byte) error {
	return errUnimplementedHandler
}
