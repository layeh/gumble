package gumble

import (
	"bytes"
	"errors"
	"net"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/layeh/gopus"
	"github.com/layeh/gumble/gumble/MumbleProto"
	"github.com/layeh/gumble/gumble/varint"
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

func handleVersion(c *Client, buffer []byte) error {
	var packet MumbleProto.Version
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	c.server.version = parseVersion(&packet)
	return nil
}

func handleUdpTunnel(c *Client, buffer []byte) error {
	reader := bytes.NewReader(buffer)
	var bytesRead int64

	var audioType byte
	var audioTarget byte
	var user *User
	var audioLength int

	// Header byte
	typeTarget, err := varint.ReadByte(reader)
	if err != nil {
		return err
	}
	audioType = (typeTarget >> 5) & 0x7
	audioTarget = typeTarget & 0x1F
	// Opus only
	if audioType != 4 {
		return errInvalidProtobuf
	}
	bytesRead++

	// Session
	session, n, err := varint.ReadFrom(reader)
	if err != nil {
		return err
	}
	user = c.users[uint(session)]
	if user == nil {
		return errInvalidProtobuf
	}
	bytesRead += n

	// Sequence
	sequence, n, err := varint.ReadFrom(reader)
	if err != nil {
		return err
	}
	bytesRead += n

	// Length
	length, n, err := varint.ReadFrom(reader)
	if err != nil {
		return err
	}
	audioLength = int(length)
	if audioLength > reader.Len() {
		return errInvalidProtobuf
	}
	bytesRead += n

	opus := buffer[bytesRead : bytesRead+int64(audioLength)]
	pcm, err := user.decoder.Decode(opus, AudioMaximumFrameSize, false)
	if err != nil {
		return err
	}
	event := AudioPacketEvent{
		Client: c,
		AudioPacket: AudioPacket{
			Sender:   user,
			Target:   int(audioTarget),
			Sequence: int(sequence),
			Pcm:      pcm,
		},
	}
	if listener := c.config.AudioListener; listener != nil {
		listener.OnAudioPacket(&event)
	}
	return nil
}

func handleAuthenticate(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handlePing(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleReject(c *Client, buffer []byte) error {
	var packet MumbleProto.Reject
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Type != nil {
		c.disconnectEvent.Type = DisconnectType(*packet.Type)
	}
	if packet.Reason != nil {
		c.disconnectEvent.String = *packet.Reason
	}
	c.connection.Close()
	return nil
}

func handleServerSync(c *Client, buffer []byte) error {
	var packet MumbleProto.ServerSync
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := ConnectEvent{
		Client: c,
	}

	if packet.Session != nil {
		c.self = c.users.BySession(uint(*packet.Session))
	}
	if packet.WelcomeText != nil {
		event.WelcomeMessage = *packet.WelcomeText
	}
	if packet.MaxBandwidth != nil {
		event.MaximumBitrate = int(*packet.MaxBandwidth)
	}
	c.state = StateSynced

	if listener := c.config.Listener; listener != nil {
		listener.OnConnect(&event)
	}
	return nil
}

func handleChannelRemove(c *Client, buffer []byte) error {
	var packet MumbleProto.ChannelRemove
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.ChannelId == nil {
		return errIncompleteProtobuf
	}
	var channel *Channel
	{
		channelID := uint(*packet.ChannelId)
		channel = c.channels.ByID(channelID)
		if channel == nil {
			return errInvalidProtobuf
		}
		c.channels.delete(channelID)
		if parent := channel.parent; parent != nil {
			channel.parent.children.delete(uint(channel.id))
		}
	}

	if c.state == StateSynced {
		event := ChannelChangeEvent{
			Client:  c,
			Type:    ChannelChangeRemoved,
			Channel: channel,
		}
		if listener := c.config.Listener; listener != nil {
			listener.OnChannelChange(&event)
		}
	}
	return nil
}

func handleChannelState(c *Client, buffer []byte) error {
	var packet MumbleProto.ChannelState
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.ChannelId == nil {
		return errIncompleteProtobuf
	}
	event := ChannelChangeEvent{
		Client: c,
	}
	var channel *Channel
	channelID := uint(*packet.ChannelId)
	if !c.channels.Exists(channelID) {
		channel = c.channels.create(channelID)
		channel.client = c

		event.Type |= ChannelChangeCreated
	} else {
		channel = c.channels.ByID(channelID)
	}
	event.Channel = channel
	if packet.Parent != nil {
		if channel.parent != nil {
			channel.parent.children.delete(channelID)
		}
		newParent := c.channels.ByID(uint(*packet.Parent))
		if newParent != channel.parent {
			event.Type |= ChannelChangeMoved
		}
		channel.parent = newParent
		if channel.parent != nil {
			channel.parent.children[uint(channel.id)] = channel
		}
	}
	if packet.Name != nil {
		newName := *packet.Name
		if newName != channel.name {
			event.Type |= ChannelChangeName
		}
		channel.name = newName
	}
	if packet.Description != nil {
		newDescription := *packet.Description
		if newDescription != channel.description {
			event.Type |= ChannelChangeDescription
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
		event.Type |= ChannelChangeDescription
		channel.descriptionHash = packet.DescriptionHash
		channel.description = ""
	}

	if c.state == StateSynced {
		if listener := c.config.Listener; listener != nil {
			listener.OnChannelChange(&event)
		}
	}
	return nil
}

func handleUserRemove(c *Client, buffer []byte) error {
	var packet MumbleProto.UserRemove
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	event := UserChangeEvent{
		Client: c,
		Type:   UserChangeDisconnected,
	}
	{
		session := uint(*packet.Session)
		event.User = c.users.BySession(session)
		if event.User == nil {
			return errInvalidProtobuf
		}
		if event.User.channel != nil {
			event.User.channel.users.delete(session)
		}
		c.users.delete(session)
	}
	if packet.Actor != nil {
		event.Actor = c.users.BySession(uint(*packet.Actor))
		if event.Actor == nil {
			return errInvalidProtobuf
		}
		event.Type |= UserChangeKicked
	}
	if packet.Reason != nil {
		event.String = *packet.Reason
	}
	if packet.Ban != nil && *packet.Ban {
		event.Type |= UserChangeBanned
	}

	if c.state == StateSynced {
		if listener := c.config.Listener; listener != nil {
			listener.OnUserChange(&event)
		}
	}
	return nil
}

func handleUserState(c *Client, buffer []byte) error {
	var packet MumbleProto.UserState
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	event := UserChangeEvent{
		Client: c,
	}
	var user, actor *User
	{
		session := uint(*packet.Session)
		if !c.users.Exists(session) {
			user = c.users.create(session)
			user.channel = c.channels.ByID(0)
			user.client = c

			event.Type |= UserChangeConnected

			decoder, _ := gopus.NewDecoder(AudioSampleRate, 1)
			user.decoder = decoder

			if user.channel == nil {
				return errInvalidProtobuf
			}
			event.Type |= UserChangeChannel
			user.channel.users[session] = user
		} else {
			user = c.users.BySession(session)
		}
	}
	event.User = user
	if packet.Actor != nil {
		actor = c.users.BySession(uint(*packet.Actor))
		if actor == nil {
			return errInvalidProtobuf
		}
		event.Actor = actor
	}
	if packet.Name != nil {
		newName := *packet.Name
		if newName != user.name {
			event.Type |= UserChangeName
		}
		user.name = newName
	}
	if packet.UserId != nil {
		user.userID = *packet.UserId
	}
	if packet.ChannelId != nil {
		if user.channel != nil {
			user.channel.users.delete(user.Session())
		}
		newChannel := c.channels.ByID(uint(*packet.ChannelId))
		if newChannel == nil {
			return errInvalidProtobuf
		}
		if newChannel != user.channel {
			event.Type |= UserChangeChannel
			user.channel = newChannel
		}
		user.channel.users[user.Session()] = user
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
			event.Type |= UserChangeComment
		}
		user.comment = newComment
		user.commentHash = nil
	}
	if packet.Hash != nil {
		user.hash = *packet.Hash
	}
	if packet.CommentHash != nil {
		event.Type |= UserChangeComment
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

	if c.state == StateSynced {
		if listener := c.config.Listener; listener != nil {
			listener.OnUserChange(&event)
		}
	}
	return nil
}

func handleBanList(c *Client, buffer []byte) error {
	var packet MumbleProto.BanList
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := BanListEvent{
		Client:  c,
		BanList: make(BanList, 0, len(packet.Bans)),
	}

	for _, banPacket := range packet.Bans {
		ban := &Ban{
			address: net.IP(banPacket.Address),
		}
		if banPacket.Mask != nil {
			size := net.IPv4len * 8
			if len(ban.address) == net.IPv6len {
				size = net.IPv6len * 8
			}
			ban.mask = net.CIDRMask(int(*banPacket.Mask), size)
		}
		if banPacket.Name != nil {
			ban.name = *banPacket.Name
		}
		if banPacket.Hash != nil {
			ban.hash = *banPacket.Hash
		}
		if banPacket.Reason != nil {
			ban.reason = *banPacket.Reason
		}
		if banPacket.Start != nil {
			ban.start, _ = time.Parse(time.RFC3339, *banPacket.Start)
		}
		if banPacket.Duration != nil {
			ban.duration = time.Duration(*banPacket.Duration) * time.Second
		}
		event.BanList = append(event.BanList, ban)
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnBanList(&event)
	}
	return nil
}

func handleTextMessage(c *Client, buffer []byte) error {
	var packet MumbleProto.TextMessage
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := TextMessageEvent{
		Client: c,
	}
	if packet.Actor != nil {
		event.Sender = c.users.BySession(uint(*packet.Actor))
	}
	if packet.Session != nil {
		event.Users = make([]*User, 0, len(packet.Session))
		for _, session := range packet.Session {
			if user := c.users.BySession(uint(session)); user != nil {
				event.Users = append(event.Users, user)
			}
		}
	}
	if packet.ChannelId != nil {
		event.Channels = make([]*Channel, 0, len(packet.ChannelId))
		for _, id := range packet.ChannelId {
			if channel := c.channels.ByID(uint(id)); channel != nil {
				event.Channels = append(event.Channels, channel)
			}
		}
	}
	if packet.TreeId != nil {
		event.Trees = make([]*Channel, 0, len(packet.TreeId))
		for _, id := range packet.TreeId {
			if channel := c.channels.ByID(uint(id)); channel != nil {
				event.Trees = append(event.Trees, channel)
			}
		}
	}
	if packet.Message != nil {
		event.Message = *packet.Message
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnTextMessage(&event)
	}
	return nil
}

func handlePermissionDenied(c *Client, buffer []byte) error {
	var packet MumbleProto.PermissionDenied
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Type == nil || *packet.Type == MumbleProto.PermissionDenied_H9K {
		return errInvalidProtobuf
	}

	event := PermissionDeniedEvent{
		Client: c,
		Type:   PermissionDeniedType(*packet.Type),
	}
	if packet.Reason != nil {
		event.String = *packet.Reason
	}
	if packet.Name != nil {
		event.String = *packet.Name
	}
	if packet.Session != nil {
		event.User = c.users.BySession(uint(*packet.Session))
		if event.User == nil {
			return errInvalidProtobuf
		}
	}
	if packet.ChannelId != nil {
		event.Channel = c.channels.ByID(uint(*packet.ChannelId))
		if event.Channel == nil {
			return errInvalidProtobuf
		}
	}
	if packet.Permission != nil {
		event.Permission = Permission(*packet.Permission)
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnPermissionDenied(&event)
	}
	return nil
}

func handleAcl(c *Client, buffer []byte) error {
	var packet MumbleProto.ACL
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := AclEvent{
		Client: c,
		Acl:    &Acl{},
	}
	if packet.ChannelId != nil {
		event.Acl.channel = c.channels.ByID(uint(*packet.ChannelId))
	}

	if packet.Groups != nil {
		event.Acl.groups = make([]*AclGroup, 0, len(packet.Groups))
		for _, group := range packet.Groups {
			event.Acl.groups = append(event.Acl.groups, &AclGroup{
				name: *group.Name,
			})
		}
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnAcl(&event)
	}
	return nil
}

func handleQueryUsers(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleCryptSetup(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleContextActionModify(c *Client, buffer []byte) error {
	var packet MumbleProto.ContextActionModify
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Action == nil || packet.Operation == nil {
		return errInvalidProtobuf
	}

	event := ContextActionChangeEvent{
		Client: c,
	}

	switch *packet.Operation {
	case MumbleProto.ContextActionModify_Add:
		if c.contextActions.Exists(*packet.Action) {
			return nil
		}
		event.Type = ContextActionAdd
		contextAction := c.contextActions.create(*packet.Action)
		if packet.Text != nil {
			contextAction.label = *packet.Text
		}
		if packet.Context != nil {
			contextAction.contextType = ContextActionType(*packet.Context)
		}
		event.ContextAction = contextAction
	case MumbleProto.ContextActionModify_Remove:
		if !c.contextActions.Exists(*packet.Action) {
			return nil
		}
		event.Type = ContextActionRemove
		contextAction := c.contextActions[*packet.Action]
		c.contextActions.delete(*packet.Action)
		event.ContextAction = contextAction
	default:
		return errInvalidProtobuf
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnContextActionChange(&event)
	}
	return nil
}

func handleContextAction(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleUserList(c *Client, buffer []byte) error {
	var packet MumbleProto.UserList
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := UserListEvent{
		Client:   c,
		UserList: make(RegisteredUsers, 0, len(packet.Users)),
	}

	for _, user := range packet.Users {
		registeredUser := &RegisteredUser{
			userID: *user.UserId,
		}
		if user.Name != nil {
			registeredUser.name = *user.Name
		}
		event.UserList = append(event.UserList, registeredUser)
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnUserList(&event)
	}
	return nil
}

func handleVoiceTarget(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handlePermissionQuery(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleCodecVersion(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleUserStats(c *Client, buffer []byte) error {
	var packet MumbleProto.UserStats
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	user := c.users.BySession(uint(*packet.Session))
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
		Client: c,
		Type:   UserChangeStats,
		User:   user,
	}

	if listener := c.config.Listener; listener != nil {
		listener.OnUserChange(&event)
	}
	return nil
}

func handleRequestBlob(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleServerConfig(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handleSuggestConfig(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}
