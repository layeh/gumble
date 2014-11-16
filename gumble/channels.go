package gumble

type Channels map[uint]*Channel

// Create adds a new channel with the given id to the collection. If a channel
// with the given id already exists, it is overwritten.
func (c *Channels) Create(id uint) *Channel {
	channel := &Channel{
		id: uint32(id),
	}
	(*c)[id] = channel
	return channel
}

// ById returns a pointer to the channel with the given id, null if no channel
// exists with the given id.
func (c *Channels) ById(id uint) *Channel {
	return (*c)[id]
}

// Exists returns true if a channel with the given id exists in the collection.
func (c *Channels) Exists(id uint) bool {
	_, ok := (*c)[id]
	return ok
}

// Delete removes the channel with the given id from the collection.
func (c *Channels) Delete(id uint) {
	delete(*c, id)
}

// TODO: document me
func (c *Channels) Find(names ...string) *Channel {
	root := (*c)[0]
	if names == nil || root == nil {
		return nil
	}
	return root.Find(names...)
}
