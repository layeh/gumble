package gumble

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/layeh/gopus"
	"github.com/layeh/gumble/gumble/MumbleProto"
	"github.com/layeh/gumble/gumble/varint"
)

type handlerFunc func(*Client, []byte) error

var (
	errUnimplementedHandler = errors.New("the handler has not been implemented")
	errIncompleteProtobuf   = errors.New("protobuf message is missing a required field")
	errInvalidProtobuf      = errors.New("protobuf message has an invalid field")
	errUnsupportedAudio     = errors.New("unsupported audio codec")
)

var handlers = map[uint16]handlerFunc{
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
	13: handleACL,
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
		return errUnsupportedAudio
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
	// Opus audio packets set the 13th bit in the size field as the terminator.
	audioLength = int(length) &^ 0x2000
	if audioLength > reader.Len() {
		return errInvalidProtobuf
	}
	audioLength64 := int64(audioLength)
	bytesRead += n

	opus := buffer[bytesRead : bytesRead+audioLength64]
	pcm, err := user.decoder.Decode(opus, AudioMaximumFrameSize, false)
	if err != nil {
		return err
	}
	event := AudioPacketEvent{
		Client: c,
	}
	event.AudioPacket.Sender = user
	event.AudioPacket.Target = int(audioTarget)
	event.AudioPacket.Sequence = int(sequence)
	event.AudioPacket.PositionalAudioBuffer.AudioBuffer = pcm

	reader.Seek(audioLength64, 1)
	binary.Read(reader, binary.LittleEndian, &event.AudioPacket.PositionalAudioBuffer.X)
	binary.Read(reader, binary.LittleEndian, &event.AudioPacket.PositionalAudioBuffer.Y)
	binary.Read(reader, binary.LittleEndian, &event.AudioPacket.PositionalAudioBuffer.Z)

	c.audioListeners.OnAudioPacket(&event)
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
		c.self = c.users[uint(*packet.Session)]
	}
	if packet.WelcomeText != nil {
		event.WelcomeMessage = *packet.WelcomeText
	}
	if packet.MaxBandwidth != nil {
		event.MaximumBitrate = int(*packet.MaxBandwidth)
	}
	c.state = StateSynced

	c.listeners.OnConnect(&event)
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
		channel = c.channels[channelID]
		if channel == nil {
			return errInvalidProtobuf
		}
		delete(c.channels, channelID)
		delete(c.permissions, channelID)
		if parent := channel.parent; parent != nil {
			delete(channel.parent.children, uint(channel.id))
		}
	}

	if c.state == StateSynced {
		event := ChannelChangeEvent{
			Client:  c,
			Type:    ChannelChangeRemoved,
			Channel: channel,
		}
		c.listeners.OnChannelChange(&event)
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
	channelID := uint(*packet.ChannelId)
	channel := c.channels[channelID]
	if channel == nil {
		channel = c.channels.create(channelID)
		channel.client = c

		event.Type |= ChannelChangeCreated
	}
	event.Channel = channel
	if packet.Parent != nil {
		if channel.parent != nil {
			delete(channel.parent.children, channelID)
		}
		newParent := c.channels[uint(*packet.Parent)]
		if newParent != channel.parent {
			event.Type |= ChannelChangeMoved
		}
		channel.parent = newParent
		if channel.parent != nil {
			channel.parent.children[uint(channel.id)] = channel
		}
	}
	if packet.Name != nil {
		if *packet.Name != channel.name {
			event.Type |= ChannelChangeName
		}
		channel.name = *packet.Name
	}
	if packet.Description != nil {
		if *packet.Description != channel.description {
			event.Type |= ChannelChangeDescription
		}
		channel.description = *packet.Description
		channel.descriptionHash = nil
	}
	if packet.Temporary != nil {
		channel.temporary = *packet.Temporary
	}
	if packet.Position != nil {
		if *packet.Position != channel.position {
			event.Type |= ChannelChangePosition
		}
		channel.position = *packet.Position
	}
	if packet.DescriptionHash != nil {
		event.Type |= ChannelChangeDescription
		channel.descriptionHash = packet.DescriptionHash
		channel.description = ""
	}

	if c.state == StateSynced {
		c.listeners.OnChannelChange(&event)
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
		event.User = c.users[session]
		if event.User == nil {
			return errInvalidProtobuf
		}
		if event.User.channel != nil {
			delete(event.User.channel.users, session)
		}
		delete(c.users, session)
	}
	if packet.Actor != nil {
		event.Actor = c.users[uint(*packet.Actor)]
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
		c.listeners.OnUserChange(&event)
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
		user = c.users[session]
		if user == nil {
			user = c.users.create(session)
			user.channel = c.channels[0]
			user.client = c

			event.Type |= UserChangeConnected

			decoder, _ := gopus.NewDecoder(AudioSampleRate, 1)
			user.decoder = decoder

			if user.channel == nil {
				return errInvalidProtobuf
			}
			event.Type |= UserChangeChannel
			user.channel.users[session] = user
		}
	}
	event.User = user
	if packet.Actor != nil {
		actor = c.users[uint(*packet.Actor)]
		if actor == nil {
			return errInvalidProtobuf
		}
		event.Actor = actor
	}
	if packet.Name != nil {
		if *packet.Name != user.name {
			event.Type |= UserChangeName
		}
		user.name = *packet.Name
	}
	if packet.UserId != nil {
		if *packet.UserId != user.userID && !event.Type.Has(UserChangeConnected) {
			if *packet.UserId != math.MaxUint32 {
				event.Type |= UserChangeRegistered
				user.userID = *packet.UserId
			} else {
				event.Type |= UserChangeUnregistered
				user.userID = 0
			}
		} else {
			user.userID = *packet.UserId
		}
	}
	if packet.ChannelId != nil {
		if user.channel != nil {
			delete(user.channel.users, user.Session())
		}
		newChannel := c.channels[uint(*packet.ChannelId)]
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
		if *packet.Mute != user.mute {
			event.Type |= UserChangeAudio
		}
		user.mute = *packet.Mute
	}
	if packet.Deaf != nil {
		if *packet.Deaf != user.deaf {
			event.Type |= UserChangeAudio
		}
		user.deaf = *packet.Deaf
	}
	if packet.Suppress != nil {
		if *packet.Suppress != user.suppress {
			event.Type |= UserChangeAudio
		}
		user.suppress = *packet.Suppress
	}
	if packet.SelfMute != nil {
		if *packet.SelfMute != user.selfMute {
			event.Type |= UserChangeAudio
		}
		user.selfMute = *packet.SelfMute
	}
	if packet.SelfDeaf != nil {
		if *packet.SelfDeaf != user.selfDeaf {
			event.Type |= UserChangeAudio
		}
		user.selfDeaf = *packet.SelfDeaf
	}
	if packet.Texture != nil {
		event.Type |= UserChangeTexture
		user.texture = packet.Texture
		user.textureHash = nil
	}
	if packet.Comment != nil {
		if *packet.Comment != user.comment {
			event.Type |= UserChangeComment
		}
		user.comment = *packet.Comment
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
		event.Type |= UserChangeTexture
		user.textureHash = packet.TextureHash
		user.texture = nil
	}
	if packet.PrioritySpeaker != nil {
		if *packet.PrioritySpeaker != user.prioritySpeaker {
			event.Type |= UserChangePrioritySpeaker
		}
		user.prioritySpeaker = *packet.PrioritySpeaker
	}
	if packet.Recording != nil {
		if *packet.Recording != user.recording {
			event.Type |= UserChangeRecording
		}
		user.recording = *packet.Recording
	}

	if c.state == StateSynced {
		c.listeners.OnUserChange(&event)
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

	c.listeners.OnBanList(&event)
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
		event.Sender = c.users[uint(*packet.Actor)]
	}
	if packet.Session != nil {
		event.Users = make([]*User, 0, len(packet.Session))
		for _, session := range packet.Session {
			if user := c.users[uint(session)]; user != nil {
				event.Users = append(event.Users, user)
			}
		}
	}
	if packet.ChannelId != nil {
		event.Channels = make([]*Channel, 0, len(packet.ChannelId))
		for _, id := range packet.ChannelId {
			if channel := c.channels[uint(id)]; channel != nil {
				event.Channels = append(event.Channels, channel)
			}
		}
	}
	if packet.TreeId != nil {
		event.Trees = make([]*Channel, 0, len(packet.TreeId))
		for _, id := range packet.TreeId {
			if channel := c.channels[uint(id)]; channel != nil {
				event.Trees = append(event.Trees, channel)
			}
		}
	}
	if packet.Message != nil {
		event.Message = *packet.Message
	}

	c.listeners.OnTextMessage(&event)
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
		event.User = c.users[uint(*packet.Session)]
		if event.User == nil {
			return errInvalidProtobuf
		}
	}
	if packet.ChannelId != nil {
		event.Channel = c.channels[uint(*packet.ChannelId)]
		if event.Channel == nil {
			return errInvalidProtobuf
		}
	}
	if packet.Permission != nil {
		event.Permission = Permission(*packet.Permission)
	}

	c.listeners.OnPermissionDenied(&event)
	return nil
}

func handleACL(c *Client, buffer []byte) error {
	var packet MumbleProto.ACL
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	acl := &ACL{
		inherits: packet.GetInheritAcls(),
	}
	if packet.ChannelId == nil {
		return errInvalidProtobuf
	}
	acl.channel = c.channels[uint(*packet.ChannelId)]
	if acl.channel == nil {
		return errInvalidProtobuf
	}

	if packet.Groups != nil {
		acl.groups = make([]*ACLGroup, 0, len(packet.Groups))
		for _, group := range packet.Groups {
			aclGroup := &ACLGroup{
				name:         *group.Name,
				inherited:    group.GetInherited(),
				inheritUsers: group.GetInherit(),
				inheritable:  group.GetInheritable(),
			}
			if group.Add != nil {
				aclGroup.usersAdd = make(map[uint32]*ACLUser)
				for _, userID := range group.Add {
					aclGroup.usersAdd[userID] = &ACLUser{
						userID: userID,
					}
				}
			}
			if group.Remove != nil {
				aclGroup.usersRemove = make(map[uint32]*ACLUser)
				for _, userID := range group.Remove {
					aclGroup.usersRemove[userID] = &ACLUser{
						userID: userID,
					}
				}
			}
			if group.InheritedMembers != nil {
				aclGroup.usersInherited = make(map[uint32]*ACLUser)
				for _, userID := range group.InheritedMembers {
					aclGroup.usersInherited[userID] = &ACLUser{
						userID: userID,
					}
				}
			}
			acl.groups = append(acl.groups, aclGroup)
		}
	}
	if packet.Acls != nil {
		acl.rules = make([]*ACLRule, 0, len(packet.Acls))
		for _, rule := range packet.Acls {
			aclRule := &ACLRule{
				appliesCurrent:  rule.GetApplyHere(),
				appliesChildren: rule.GetApplySubs(),
				inherited:       rule.GetInherited(),
				granted:         Permission(rule.GetGrant()),
				denied:          Permission(rule.GetDeny()),
			}
			if rule.UserId != nil {
				aclRule.user = &ACLUser{
					userID: *rule.UserId,
				}
			} else if rule.Group != nil {
				var group *ACLGroup
				for _, g := range acl.groups {
					if g.name == *rule.Group {
						group = g
						break
					}
				}
				if group == nil {
					group = &ACLGroup{
						name: *rule.Group,
					}
				}
				aclRule.group = group
			}
			acl.rules = append(acl.rules, aclRule)
		}
	}
	c.tmpACL = acl
	return nil
}

func handleQueryUsers(c *Client, buffer []byte) error {
	var packet MumbleProto.QueryUsers
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	acl := c.tmpACL
	c.tmpACL = nil

	userMap := make(map[uint32]string)
	for i := 0; i < len(packet.Ids) && i < len(packet.Names); i++ {
		userMap[packet.Ids[i]] = packet.Names[i]
	}

	for _, group := range acl.groups {
		for _, user := range group.usersAdd {
			user.name = userMap[user.userID]
		}
		for _, user := range group.usersRemove {
			user.name = userMap[user.userID]
		}
		for _, user := range group.usersInherited {
			user.name = userMap[user.userID]
		}
	}
	for _, rule := range acl.rules {
		if rule.user != nil {
			rule.user.name = userMap[rule.user.userID]
		}
	}

	event := ACLEvent{
		Client: c,
		ACL:    acl,
	}
	c.listeners.OnACL(&event)
	return nil
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
		if ca := c.contextActions[*packet.Action]; ca != nil {
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
		contextAction := c.contextActions[*packet.Action]
		if contextAction == nil {
			return nil
		}
		event.Type = ContextActionRemove
		delete(c.contextActions, *packet.Action)
		event.ContextAction = contextAction
	default:
		return errInvalidProtobuf
	}

	c.listeners.OnContextActionChange(&event)
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

	c.listeners.OnUserList(&event)
	return nil
}

func handleVoiceTarget(c *Client, buffer []byte) error {
	return errUnimplementedHandler
}

func handlePermissionQuery(c *Client, buffer []byte) error {
	var packet MumbleProto.PermissionQuery
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Flush != nil && *packet.Flush {
		oldPermissions := c.permissions
		c.permissions = make(map[uint]*Permission)
		for channelID := range oldPermissions {
			channel := c.channels[channelID]
			event := ChannelChangeEvent{
				Client:  c,
				Type:    ChannelChangePermission,
				Channel: channel,
			}
			c.listeners.OnChannelChange(&event)
		}
	}
	if packet.ChannelId != nil {
		channel := c.channels[uint(*packet.ChannelId)]
		if packet.Permissions != nil {
			p := Permission(*packet.Permissions)
			c.permissions[channel.ID()] = &p
			event := ChannelChangeEvent{
				Client:  c,
				Type:    ChannelChangePermission,
				Channel: channel,
			}
			c.listeners.OnChannelChange(&event)
		}
	}
	return nil
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
	user := c.users[uint(*packet.Session)]
	if user == nil {
		return errInvalidProtobuf
	}

	if user.stats == nil {
		user.stats = &UserStats{}
	}
	*user.stats = UserStats{
		user: user,
	}
	stats := user.stats

	if packet.Version != nil {
		stats.version = parseVersion(packet.Version)
	}
	if packet.Onlinesecs != nil {
		stats.connected = time.Now().Add(time.Duration(*packet.Onlinesecs) * -time.Second)
	}
	if packet.Idlesecs != nil {
		stats.idle = time.Duration(*packet.Idlesecs) * time.Second
	}
	if packet.Bandwidth != nil {
		stats.bandwidth = int(*packet.Bandwidth)
	}
	if packet.Address != nil {
		stats.ip = net.IP(packet.Address)
	}
	if packet.Certificates != nil {
		stats.certificates = make([]*x509.Certificate, 0, len(packet.Certificates))
		for _, data := range packet.Certificates {
			if data != nil {
				if cert, err := x509.ParseCertificate(data); err == nil {
					stats.certificates = append(stats.certificates, cert)
				}
			}
		}
	}
	if packet.Opus != nil {
		stats.opus = *packet.Opus
	}

	event := UserChangeEvent{
		Client: c,
		Type:   UserChangeStats,
		User:   user,
	}

	c.listeners.OnUserChange(&event)
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
