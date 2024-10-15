package controller_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/namlh/vulcanLabsOA/config"
	"github.com/namlh/vulcanLabsOA/controller"
	"github.com/namlh/vulcanLabsOA/manager"
	"github.com/namlh/vulcanLabsOA/testing/assert"
)

func TestRoomController_ListAvailableSeats(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	logger := slog.Default()
	cfg := config.Room{
		NumRows:     4,
		NumCols:     4,
		MinDistance: 3,
	}

	groupManager := manager.NewGroupManager([]string{"abc"})
	roomManager := manager.NewRoomManager(logger, &cfg, groupManager)
	ctrl := controller.NewRoomController(logger, roomManager)

	srv := httptest.NewServer(http.HandlerFunc(ctrl.ListAvailableSeats))
	t.Cleanup(srv.Close)

	testcases := []struct {
		name       string
		groupID    string
		assertFunc func(t *testing.T, resp *http.Response)
	}{
		{
			name: "success/list all",
			assertFunc: func(t *testing.T, resp *http.Response) {
				buf, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				expect := `{"code":0,"message":"Success","data":{"abc":[[0,0],[0,1],[0,2],[0,3],[1,0],[1,1],[1,2],[1,3],[2,0],[2,1],[2,2],[2,3],[3,0],[3,1],[3,2],[3,3]]}}`
				assert.Equal(t, expect, string(buf))
			},
		},
		{
			name:    "success/filter by group_id",
			groupID: "abc",
			assertFunc: func(t *testing.T, resp *http.Response) {
				buf, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				expect := `{"code":0,"message":"Success","data":{"abc":[[0,0],[0,1],[0,2],[0,3],[1,0],[1,1],[1,2],[1,3],[2,0],[2,1],[2,2],[2,3],[3,0],[3,1],[3,2],[3,3]]}}`
				assert.Equal(t, expect, string(buf))
			},
		},
		{
			name:    "fail/group_id not found",
			groupID: "xyz",
			assertFunc: func(t *testing.T, resp *http.Response) {
				buf, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				expect := `{"code":1,"message":"Invalid parameters","details":{"group_id":"group_id not found"}}`
				assert.Equal(t, expect, string(buf))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req, err := http.NewRequestWithContext(ctx, "", srv.URL, nil)
			assert.NoError(t, err)
			if tc.groupID != "" {
				q := req.URL.Query()
				q.Set("group_id", tc.groupID)
				req.URL.RawQuery = q.Encode()
			}

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			if tc.assertFunc != nil {
				tc.assertFunc(t, resp)
			}
		})
	}
}

func TestRoomController_ReserveSeats(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	logger := slog.Default()
	cfg := config.Room{
		NumRows:     4,
		NumCols:     4,
		MinDistance: 3,
	}

	groupManager := manager.NewGroupManager([]string{"abc", "xyz"})
	roomManager := manager.NewRoomManager(logger, &cfg, groupManager)
	ctrl := controller.NewRoomController(logger, roomManager)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/seats/reservation", ctrl.ReserveSeats)
	mux.HandleFunc("/api/available-seats", ctrl.ListAvailableSeats)

	testcases := []struct {
		name       string
		req        func() io.Reader
		assertFunc func(t *testing.T, url string, resp *http.Response)
	}{
		{
			name: "name",
			req: func() io.Reader {
				rawReq := `{"seats_reservation":[{"group_id":"abc","position":[0,1]}]}`
				return strings.NewReader(rawReq)
			},
			assertFunc: func(t *testing.T, url string, resp *http.Response) {
				buf, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				expect := `{"code":0,"message":"Success"}`
				assert.Equal(t, expect, string(buf))

				req, err := http.NewRequestWithContext(ctx, "POST", url+"/api/available-seats", nil)
				assert.NoError(t, err)
				listSeatsResp, err := http.DefaultClient.Do(req)
				assert.NoError(t, err)
				t.Cleanup(func() {
					_ = listSeatsResp.Body.Close()
				})

				buf, err = io.ReadAll(listSeatsResp.Body)
				assert.NoError(t, err)
				expect = `{"code":0,"message":"Success","data":{"abc":[[0,0],[0,2],[0,3],[1,0],[1,1],[1,2],[1,3],[2,0],[2,1],[2,2],[2,3],[3,0],[3,1],[3,2],[3,3]],"xyz":[[1,3],[2,0],[2,2],[2,3],[3,0],[3,1],[3,2],[3,3]]}}`
				assert.Equal(t, expect, string(buf))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			req, err := http.NewRequestWithContext(ctx, "POST", srv.URL+"/api/seats/reservation", tc.req())
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			if tc.assertFunc != nil {
				tc.assertFunc(t, srv.URL, resp)
			}
		})
	}
}

func TestRoomController_CancelSeats(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	logger := slog.Default()
	cfg := config.Room{
		NumRows:     4,
		NumCols:     4,
		MinDistance: 3,
	}

	groupManager := manager.NewGroupManager([]string{"abc", "xyz"})
	roomManager := manager.NewRoomManager(logger, &cfg, groupManager)
	ctrl := controller.NewRoomController(logger, roomManager)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/seats/reservation", ctrl.ReserveSeats)
	mux.HandleFunc("POST /api/seats/cancellation", ctrl.CancelSeats)
	mux.HandleFunc("/api/available-seats", ctrl.ListAvailableSeats)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	{
		var reqBody io.Reader = strings.NewReader(`{"seats_reservation":[{"group_id":"abc","position":[0,1]}]}`)
		req, err := http.NewRequestWithContext(ctx, "POST", srv.URL+"/api/seats/reservation", reqBody)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })

		buf, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"code":0,"message":"Success"}`, string(buf))
	}

	{
		var reqBody io.Reader = strings.NewReader(`{"seats_cancellation":[{"position":[0,1]}]}`)
		req, err := http.NewRequestWithContext(ctx, "POST", srv.URL+"/api/seats/cancellation", reqBody)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })

		buf, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"code":0,"message":"Success"}`, string(buf))
	}

	{
		req, err := http.NewRequestWithContext(ctx, "POST", srv.URL+"/api/available-seats", nil)
		assert.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })

		buf, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		arr := make([][2]int, 0)
		for i := 0; i < cfg.NumRows; i++ {
			for j := 0; j < cfg.NumCols; j++ {
				arr = append(arr, [2]int{i, j})
			}
		}
		expect, _ := json.Marshal(controller.NewSuccessResponse(map[string][][2]int{
			"abc": arr,
			"xyz": arr,
		}))
		assert.Equal(t, string(expect), string(buf))
	}
}
