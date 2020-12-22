package app

import (
	"github.com/adnaan/users"
	"github.com/zpatrick/rbac"
)

func allowAll(_ string) rbac.Matcher {
	return func(_ string) (bool, error) {
		return true, nil
	}
}

func ifTaskOwner(t TasksContext) func(string) rbac.Matcher {
	return func(userID string) rbac.Matcher {
		return func(target string) (bool, error) {
			tsk, err := t.client.Task.Get(t.ctx, target)
			if err != nil {
				return false, err
			}

			if tsk.Owner == userID {
				return true, nil
			}

			return false, nil
		}
	}
}

func ownerRole(t TasksContext) []users.Permission {
	return []users.Permission{
		users.NewPermission("get:api:tasks", allowAll),
		users.NewPermission("post:api:tasks", allowAll),
		users.NewPermission("delete:api:tasks:*", ifTaskOwner(t))}
}
