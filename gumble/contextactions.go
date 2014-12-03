package gumble

type ContextActions map[string]*ContextAction

func (ca ContextActions) create(action string) *ContextAction {
	contextAction := &ContextAction{
		name: action,
	}
	ca[action] = contextAction
	return contextAction
}

// ByName returns a pointer to the ContextAction with the given action name.
func (ca ContextActions) ByName(action string) *ContextAction {
	return ca[action]
}

func (ca ContextActions) delete(action string) {
	delete(ca, action)
}

// Exists returns if the action with the given name exists in the collection.
func (ca ContextActions) Exists(action string) bool {
	_, ok := ca[action]
	return ok
}
