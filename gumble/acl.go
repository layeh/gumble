package gumble

//
type ACL struct {
	channel *Channel
	groups  []*ACLGroup
}

// Channel returns the channel to which the ACL is referring.
func (a *ACL) Channel() *Channel {
	return a.channel
}

// Groups return a slice of ACL groups part of the ACL.
func (a *ACL) Groups() []*ACLGroup {
	return a.groups
}

type ACLGroup struct {
	name string
}

// Name returns the ACL group name.
func (ag *ACLGroup) Name() string {
	return ag.name
}
