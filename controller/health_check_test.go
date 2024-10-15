package controller_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/namlh/vulcanLabsOA/controller"
	"github.com/namlh/vulcanLabsOA/testing/assert"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	srv := httptest.NewServer(controller.HealthCheck(slog.Default()))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(ctx, "", srv.URL, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, 200, resp.StatusCode)
	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK\n", string(buf))
}
