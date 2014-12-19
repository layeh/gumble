package gumble

// Channels is a map of server channels.
//
// When accessed through client.Channels(), it contains all channels on the
// server. When accessed through a specific channel
// (e.g. client.Channels()[0].Channels()), it contains only the children of the
// channel.
type Channels map[uint]*Channel

// create adds a new channel with the given id to the collection. If a channel
// with the given id already exists, it is overwritten.
func (c Channels) create(id uint) *Channel {
	channel := &Channel{
		id:       uint32(id),
		children: Channels{},
		users:    Users{},
	}
	c[id] = channel
	return channel
}

// ByID returns a pointer to the channel with the given id, nil if no channel
// exists with the given id.
func (c Channels) ByID(id uint) *Channel {
	return c[id]
}

// Exists returns true if a channel with the given id exists in the collection.
func (c Channels) Exists(id uint) bool {
	_, ok := c[id]
	return ok
}

// delete removes the channel with the given id from the collection.
func (c Channels) delete(id uint) {
	delete(c, id)
}

// Find returns a channel whose path (by channel name) from the server root
// channel is equal to the arguments passed. If the root channel does not
// exist, nil is returned.
func (c Channels) Find(names ...string) *Channel {
	root := c[0]
	if names == nil || root == nil {
		return root
	}
	return root.Find(names...)
}
