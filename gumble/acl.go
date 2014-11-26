package gumble

type Acl struct {
	channel *Channel
	groups  []*AclGroup
}

// Channel returns the channel to which the Acl is referring.
func (a *Acl) Channel() *Channel {
	return a.channel
}

// Groups return a slice of Acl groups part of the Acl.
func (a *Acl) Groups() []*AclGroup {
	return a.groups
}

type AclGroup struct {
	name string
}

// Name returns the Acl group name.
func (ag *AclGroup) Name() string {
	return ag.name
}
