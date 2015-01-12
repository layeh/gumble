package gumble

// ACL contains a list of ACLGroups and ACLRules linked to a channel.
type ACL struct {
	channel  *Channel
	groups   []*ACLGroup
	rules    []*ACLRule
	inherits bool
}

// Channel returns the channel to which the ACL is referring.
func (a *ACL) Channel() *Channel {
	return a.channel
}

// Groups return the ACL's groups.
func (a *ACL) Groups() []*ACLGroup {
	return a.groups
}

// Rules returns the ACL's rules.
func (a *ACL) Rules() []*ACLRule {
	return a.rules
}

// Inherits returns if the ACL inherits the parent channel's ACL.
func (a *ACL) Inherits() bool {
	return a.inherits
}

// ACLUser is a registered user who is part of or can be part of an ACL group
// or rule.
type ACLUser struct {
	userID uint32
	name   string
}

// UserID returns the user ID of the user.
func (au *ACLUser) UserID() uint {
	return uint(au.userID)
}

// Name returns the name of the user.
func (au *ACLUser) Name() string {
	return au.name
}

// ACLGroup is a named group of registered users which can be used in an
// ACLRule.
type ACLGroup struct {
	name                                  string
	inherited, inheritUsers, inheritable  bool
	usersAdd, usersRemove, usersInherited map[uint32]*ACLUser
}

// Name returns the ACL group name.
func (ag *ACLGroup) Name() string {
	return ag.name
}

// Inherited returns if the group was inherited from the parent channel's ACL.
func (ag *ACLGroup) Inherited() bool {
	return ag.inherited
}

// InheritUsers returns if group members are inherited from the parent
// channel's ACL.
func (ag *ACLGroup) InheritUsers() bool {
	return ag.inheritUsers
}

// Inheritable returns if the group can be inherited by child channels.
func (ag *ACLGroup) Inheritable() bool {
	return ag.inheritable
}

// ACLRule is a set of granted and denied permissions given to an ACLUser or
// ACLGroup.
type ACLRule struct {
	appliesCurrent, appliesChildren bool
	inherited                       bool
	user                            *ACLUser
	group                           *ACLGroup
	granted, denied                 Permission
}

// AppliesCurrent returns if the rule applies to the channel on which the rule is
// defined.
func (ar *ACLRule) AppliesCurrent() bool {
	return ar.appliesCurrent
}

// AppliesChildren returns if the rule applies to the child channels of the
// channel on which the rule is defined.
func (ar *ACLRule) AppliesChildren() bool {
	return ar.appliesChildren
}

// Inherited returns if the rule was inherited from the parent channel's ACL.
func (ar *ACLRule) Inherited() bool {
	return ar.inherited
}

// User returns the ACL user the rule applies to (can be nil).
func (ar *ACLRule) User() *ACLUser {
	return ar.user
}

// Group returns the ACL group the rule applies to (can be nil).
func (ar *ACLRule) Group() *ACLGroup {
	return ar.group
}

// Granted returns the permissions granted by the rule.
func (ar *ACLRule) Granted() Permission {
	return ar.granted
}

// Denied returns the permissions denied by the rule.
func (ar *ACLRule) Denied() Permission {
	return ar.denied
}
