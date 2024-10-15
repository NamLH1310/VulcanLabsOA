package controller_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/namlh/vulcanLabsOA/controller"
	"github.com/namlh/vulcanLabsOA/manager"
	"github.com/namlh/vulcanLabsOA/testing/assert"
)

func TestGroupController_ListGroupIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	groupManager := manager.NewGroupManager([]string{"abc", "xyz", "123"})
	ctrl := controller.NewGroupController(slog.Default(), groupManager)

	srv := httptest.NewServer(http.HandlerFunc(ctrl.ListGroupIDs))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(ctx, "", srv.URL, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, 200, resp.StatusCode)
	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"code":0,"message":"Success","data":["abc","xyz","123"]}`, string(buf))
}
