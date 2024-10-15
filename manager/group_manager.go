package manager

import "context"

type GroupManager interface {
	ListGroupIDs(ctx context.Context) []string
	HasGroupID(ctx context.Context, groupID string) bool
}

type DefaultGroupManager struct {
	groups []string
}

func NewGroupManager(groups []string) GroupManager {
	return &DefaultGroupManager{
		groups: groups,
	}
}

func (m *DefaultGroupManager) ListGroupIDs(ctx context.Context) []string {
	return m.groups
}

func (m *DefaultGroupManager) HasGroupID(ctx context.Context, groupID string) bool {
	for _, e := range m.ListGroupIDs(ctx) {
		if groupID == e {
			return true
		}
	}

	return false
}
