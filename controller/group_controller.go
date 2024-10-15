package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/namlh/vulcanLabsOA/manager"
)

type GroupController struct {
	logger  *slog.Logger
	manager manager.GroupManager
}

func NewGroupController(logger *slog.Logger, manager manager.GroupManager) *GroupController {
	return &GroupController{
		logger:  logger,
		manager: manager,
	}
}

func (c *GroupController) ListGroupIDs(w http.ResponseWriter, r *http.Request) {
	easyHandler("list group ids", w, r, c.logger, func(ctx context.Context) ([]string, error) {
		groupIDs := c.manager.ListGroupIDs(ctx)
		return groupIDs, nil
	})
}
