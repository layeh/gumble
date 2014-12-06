package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
)

type plugin struct {
	pathfs.FileSystem
	config    gumble.Config
	client    *gumble.Client
	keepAlive chan bool

	userInfo map[uint]string
}

func (p *plugin) storeUserInfo(user *gumble.User) {
	var buffer bytes.Buffer
	channelPath := strings.Join(gumbleutil.ChannelPath(user.Channel()), " -> ")

	fmt.Fprintf(&buffer, "Username:      %s\n", user.Name())
	fmt.Fprintf(&buffer, "Session ID:    %d\n", user.Session())
	fmt.Fprintf(&buffer, "Channel:       %s\n", channelPath)
	fmt.Fprintf(&buffer, "Muted:         %v\n", user.IsMuted())
	fmt.Fprintf(&buffer, "Self Muted:    %v\n", user.IsSelfMuted())
	fmt.Fprintf(&buffer, "Deafened:      %v\n", user.IsDeafened())
	fmt.Fprintf(&buffer, "Self Deafened: %v\n", user.IsSelfDeafened())
	fmt.Fprintf(&buffer, "Recording:     %v\n", user.IsRecording())
	fmt.Fprintf(&buffer, "Registered:    %v\n", user.IsRegistered())
	fmt.Fprintf(&buffer, "User ID:       %d\n", user.UserID())
	p.userInfo[user.Session()] = buffer.String()
}

func (p *plugin) OnConnect(e *gumble.ConnectEvent) {
	p.userInfo = make(map[uint]string)
	for _, user := range p.client.Users() {
		p.storeUserInfo(user)
	}
	log.Printf("connected\n")
}

func (p *plugin) OnDisconnect(e *gumble.DisconnectEvent) {
	log.Printf("disconnected\n")
	p.keepAlive <- true
}

func (p *plugin) OnUserChange(e *gumble.UserChangeEvent) {
	if e.Type.Has(gumble.UserChangeDisconnected) {
		delete(p.userInfo, e.User.Session())
	} else {
		p.storeUserInfo(e.User)
	}
}

func (p *plugin) userFromPath(name string) *gumble.User {
	basename := filepath.Base(name)
	if !strings.HasSuffix(basename, ".user") {
		return nil
	}
	username := strings.TrimSuffix(basename, ".user")
	return p.client.Users().Find(username)
}

func (p *plugin) channelFromPath(name string) *gumble.Channel {
	path := strings.Split(name, string(filepath.Separator))
	if len(path) == 1 && (path[0] == "" || path[0] == ".") {
		path = nil
	}
	for i, item := range path {
		if !strings.HasSuffix(item, ".channel") {
			return nil
		}
		path[i] = strings.TrimSuffix(item, ".channel")
	}
	return p.client.Channels().Find(path...)
}

func (p *plugin) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	path := filepath.Dir(name)
	channel := p.channelFromPath(path)
	if channel == nil {
		return nil, fuse.ENOENT
	}

	if name == "." {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}

	subUser := p.userFromPath(name)
	if subUser != nil {
		size := len(p.userInfo[subUser.Session()])
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644,
			Size: uint64(size),
		}, fuse.OK
	}

	subChannel := p.channelFromPath(name)
	if subChannel != nil {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}

	return nil, fuse.ENOENT
}

func (p *plugin) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	channel := p.channelFromPath(name)
	if channel == nil {
		return nil, fuse.ENOENT
	}

	var entries []fuse.DirEntry
	for _, channel := range channel.Channels() {
		entries = append(entries, fuse.DirEntry{
			Name: channel.Name() + ".channel",
			Mode: fuse.S_IFDIR,
		})
	}
	for _, user := range channel.Users() {
		entries = append(entries, fuse.DirEntry{
			Name: user.Name() + ".user",
			Mode: fuse.S_IFREG,
		})
	}
	return entries, fuse.OK
}

func (p *plugin) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	user := p.userFromPath(name)
	if user == nil {
		return nil, fuse.ENOENT
	}
	data := p.userInfo[user.Session()]
	return nodefs.NewDataFile([]byte(data)), fuse.OK
}

func main() {
	// flags
	server := flag.String("server", "localhost:64738", "mumble server address")
	username := flag.String("username", "gumble-fuse", "client username")
	password := flag.String("password", "", "client password")
	insecure := flag.Bool("insecure", false, "skip checking server certificate")
	path := flag.String("path", "", "path where the connection should be mounted")

	flag.Parse()

	// implementation
	p := plugin{
		keepAlive: make(chan bool),
	}

	// client
	p.client = gumble.NewClient(&p.config)
	p.config.Username = *username
	p.config.Password = *password
	p.config.Address = *server
	if *insecure {
		p.config.TLSConfig.InsecureSkipVerify = true
	}
	p.config.Listener = gumbleutil.Listener{
		Connect:    p.OnConnect,
		Disconnect: p.OnDisconnect,
		UserChange: p.OnUserChange,
	}
	if err := p.client.Connect(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	// Fuse
	p.FileSystem = pathfs.NewDefaultFileSystem()
	nfs := pathfs.NewPathNodeFs(&p, nil)
	fuseServer, _, err := nodefs.MountRoot(*path, nfs.Root(), nil)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	defer fuseServer.Unmount()
	go func() {
		fuseServer.Serve()
	}()

	interupt := make(chan os.Signal)
	signal.Notify(interupt, os.Interrupt, os.Kill)

	select {
	case <-p.keepAlive:
	case <-interupt:
	}
}
