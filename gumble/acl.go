package gumble

// ACL contains a list of ACLGroups and ACLRules linked to a channel.
type ACL struct {
	// The channel to which the ACL belongs.
	Channel *Channel
	// The ACL's groups.
	Groups []*ACLGroup
	// The ACL's rules.
	Rules []*ACLRule
	// Does the ACL inherits the parent channel's ACL?s
	Inherits bool
}

// ACLUser is a registered user who is part of or can be part of an ACL group
// or rule.
type ACLUser struct {
	// The user ID of the user.
	UserID uint32
	// The name of the user.
	Name string
}

// ACLGroup is a named group of registered users which can be used in an
// ACLRule.
type ACLGroup struct {
	// The ACL group name.
	Name string
	// Is the group inherited from the parent channel's ACL?
	Inherited bool
	// Are group members are inherited from the parent channel's ACL?
	InheritUsers bool
	// Can the group be inherited by child channels?
	Inheritable                           bool
	usersAdd, usersRemove, usersInherited map[uint32]*ACLUser
}

// ACL group names that are built-in.
const (
	ACLGroupEveryone       = "all"
	ACLGroupAuthenticated  = "auth"
	ACLGroupInsideChannel  = "in"
	ACLGroupOutsideChannel = "out"
)

// ACLRule is a set of granted and denied permissions given to an ACLUser or
// ACLGroup.
type ACLRule struct {
	// Does the rule apply to the channel in which the rule is defined?
	AppliesCurrent bool
	// Does the rule apply to the children of the channel in which the rule is
	// defined?
	AppliesChildren bool
	// Is the rule inherited from the parent channel's ACL?
	Inherited bool

	// The permissions granted by the rule.
	Granted Permission
	// The permissions denied by the rule.
	Denied Permission

	// The ACL user the rule applies to. Can be nil.
	User *ACLUser
	// The ACL group the rule applies to. Can be nil.
	Group *ACLGroup
}
