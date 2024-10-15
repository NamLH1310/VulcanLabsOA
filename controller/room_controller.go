package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/namlh/vulcanLabsOA/consts/errcode"
	"github.com/namlh/vulcanLabsOA/controller/request"
	"github.com/namlh/vulcanLabsOA/manager"
)

type RoomController struct {
	logger  *slog.Logger
	manager manager.RoomManager
}

func NewRoomController(logger *slog.Logger, manager manager.RoomManager) *RoomController {
	return &RoomController{
		logger:  logger,
		manager: manager,
	}
}

func (c *RoomController) ListAvailableSeats(w http.ResponseWriter, r *http.Request) {
	easyHandler("list available seats", w, r, c.logger, func(ctx context.Context) (map[string][]manager.Coordinate, error) {
		query := r.URL.Query()
		groupID := query.Get("group_id")

		seats, err := c.manager.ListAvailableSeats(ctx, groupID)
		if err != nil {
			if errors.Is(err, manager.ErrGroupIdNotFound) {
				return nil, AppError{
					ErrCode:    errcode.InvalidParameters,
					HttpStatus: http.StatusUnprocessableEntity,
					err: ValidationErrors{
						"group_id": "group_id not found",
					},
				}
			}

			return nil, fmt.Errorf("list available seats: %w", err)
		}

		return seats, nil
	})
}

func (c *RoomController) ReserveSeats(w http.ResponseWriter, r *http.Request) {
	easyHandler("reserve seats", w, r, c.logger, func(ctx context.Context) (any, error) {
		req, err := decodeValid[request.SeatsReservation](r)
		if err != nil {
			return nil, err
		}

		seats := make([]manager.Seat, len(req.SeatsReservation))
		for i := range req.SeatsReservation {
			seats[i] = manager.Seat{
				GroupID:    req.SeatsReservation[i].GroupID,
				Coordinate: manager.Coordinate(*req.SeatsReservation[i].Position),
			}
		}

		if err := c.manager.ReserveSeats(ctx, seats); err != nil {
			if sErr := (manager.SeatError{}); errors.As(err, &sErr) {
				appErr := AppError{
					ErrCode:    errcode.InvalidParameters,
					HttpStatus: http.StatusUnprocessableEntity,
				}

				field := ""
				switch sErr.Code {
				case manager.SeatErrorCodeOutOfBound, manager.SeatErrorCodeSeatTaken,
					manager.SeatErrorCodeInvalidDistance, manager.SeatErrorCodeDuplicatedPosition:
					field = "position"
				case manager.SeatErrorCodeGroupIDNotFound:
					field = "group_id"
				default:
					panic("unhandled default case")
				}

				appErr.err = ValidationErrors{
					field: sErr.Error(),
				}

				return nil, appErr
			}
			return nil, err
		}

		return nil, nil
	})
}

func (c *RoomController) CancelSeats(w http.ResponseWriter, r *http.Request) {
	easyHandler("cancel seats", w, r, c.logger, func(ctx context.Context) (any, error) {
		req, err := decodeValid[request.SeatsCancellation](r)
		if err != nil {
			return nil, err
		}
		seats := make([]manager.Coordinate, len(req.SeatsCancellation))
		for i := range req.SeatsCancellation {
			seats[i] = req.SeatsCancellation[i].Position
		}

		if err := c.manager.CancelSeats(ctx, seats); err != nil {
			return nil, err
		}

		return nil, nil
	})
}
