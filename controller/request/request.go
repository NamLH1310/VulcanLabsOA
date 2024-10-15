package request

import (
	"context"
	"fmt"
)

type SeatsReservation struct {
	SeatsReservation []struct {
		GroupID  string  `json:"group_id"`
		Position *[2]int `json:"position"`
	} `json:"seats_reservation"`
}

func (s SeatsReservation) Valid(_ context.Context) map[string]string {
	problems := make(map[string]string)
	for i, data := range s.SeatsReservation {
		if len(problems) > 0 {
			break
		}

		if data.Position == nil {
			problems["position"] = fmt.Sprintf("position at index %d must not be empty", i)
		} else {
			if (*data.Position)[0] < 0 || (*data.Position)[1] < 0 {
				problems["position"] = fmt.Sprintf("position at index %d must be greater than 0", i)
			}
			if data.GroupID == "" {
				problems["group_id"] = fmt.Sprintf("group_id at index %d must not be empty", i)
			}
		}

	}

	return problems
}

type SeatsCancellation struct {
	SeatsCancellation []struct {
		Position [2]int `json:"position"`
	} `json:"seats_cancellation"`
}

func (s SeatsCancellation) Valid(_ context.Context) map[string]string {
	problems := make(map[string]string)
	for i, data := range s.SeatsCancellation {
		if len(problems) > 0 {
			break
		}

		if data.Position[0] < 0 || data.Position[1] < 0 {
			problems["position"] = fmt.Sprintf("position at index %d must be greater than 0", i)
		}
	}

	return problems
}
