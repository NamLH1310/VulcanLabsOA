package manager

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"

	"github.com/namlh/vulcanLabsOA/config"
	"github.com/namlh/vulcanLabsOA/util/mathutil"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrGroupIdNotFound = Error("group id not found")
)

type SeatErrorCode int

const (
	SeatErrorCodeOutOfBound SeatErrorCode = iota + 1
	SeatErrorCodeSeatTaken
	SeatErrorCodeInvalidDistance
	SeatErrorCodeGroupIDNotFound
	SeatErrorCodeNotReserved
	SeatErrorCodeDuplicatedPosition
)

type SeatError struct {
	seat  Seat
	Code  SeatErrorCode
	index int
}

func (e SeatError) Error() string {
	switch e.Code {
	case SeatErrorCodeOutOfBound:
		return fmt.Sprintf("position [%d,%d] at index %d out of bound", e.seat.Row(), e.seat.Col(), e.index)
	case SeatErrorCodeSeatTaken:
		return fmt.Sprintf("position [%d,%d] at index %d has already been taken", e.seat.Row(), e.seat.Col(), e.index)
	case SeatErrorCodeInvalidDistance:
		return fmt.Sprintf("position [%d,%d] at index %d violate min distance constraint", e.seat.Row(), e.seat.Col(), e.index)
	case SeatErrorCodeGroupIDNotFound:
		return fmt.Sprintf("group_id %s at index %d not found", e.seat.GroupID, e.index)
	case SeatErrorCodeNotReserved:
		return fmt.Sprintf("position [%d,%d] at index %d did not get reserved", e.seat.Row(), e.seat.Col(), e.index)
	case SeatErrorCodeDuplicatedPosition:
		return fmt.Sprintf("position [%d,%d] at index %d is duplicated", e.seat.Row(), e.seat.Col(), e.index)
	}

	return ""
}

type RoomManager interface {
	ListAvailableSeats(ctx context.Context, groupID string) (map[string][]Coordinate, error)
	ReserveSeats(ctx context.Context, seats []Seat) error
	CancelSeats(ctx context.Context, seats []Coordinate) error
}

type DefaultRoomManager struct {
	logger       *slog.Logger
	cfg          *config.Room
	mu           *sync.Mutex
	reservedSeat map[int64]string
	groupManager GroupManager
}

func NewRoomManager(
	logger *slog.Logger,
	cfg *config.Room,
	groupManager GroupManager,
) RoomManager {
	return &DefaultRoomManager{
		logger:       logger,
		cfg:          cfg,
		mu:           new(sync.Mutex),
		reservedSeat: make(map[int64]string),
		groupManager: groupManager,
	}
}

func (m *DefaultRoomManager) ListAvailableSeats(ctx context.Context, groupID string) (map[string][]Coordinate, error) {
	m.mu.Lock()
	reservedSeats := maps.Clone(m.reservedSeat)
	m.mu.Unlock()

	groupIDs := []string{groupID}
	if groupID != "" && !m.groupManager.HasGroupID(ctx, groupID) {
		return nil, ErrGroupIdNotFound
	} else if groupID == "" {
		groupIDs = m.groupManager.ListGroupIDs(ctx)
	}

	availableSeatBucket := make(map[string][]Coordinate)
	for _, groupID := range groupIDs {
		availableSeatBucket[groupID] = nil
	}

	for i := int64(0); i < int64(m.cfg.NumRows)*int64(m.cfg.NumCols); i++ {
		if _, ok := reservedSeats[i]; ok {
			continue
		}

		candidateCoord := m.indexToCoordinate(i)

		for _, groupID := range groupIDs {
			isValid := true
			for k, reservedGroupID := range reservedSeats {
				a := Seat{
					GroupID:    reservedGroupID,
					Coordinate: m.indexToCoordinate(k),
				}

				b := Seat{
					GroupID:    groupID,
					Coordinate: candidateCoord,
				}

				if !m.isValidDistance(a, b) {
					isValid = false
					break
				}
			}

			if isValid {
				availableSeatBucket[groupID] = append(availableSeatBucket[groupID], candidateCoord)
			}
		}
	}

	return availableSeatBucket, nil
}

func (m *DefaultRoomManager) ReserveSeats(ctx context.Context, seats []Seat) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	numRows, numCols := m.cfg.NumRows, m.cfg.NumCols
	idxSet := make(map[int64]struct{})

	// validation
	for i, seat := range seats {
		if !m.groupManager.HasGroupID(ctx, seat.GroupID) {
			return SeatError{seat, SeatErrorCodeGroupIDNotFound, i}
		}

		if seat.Row() < 0 || seat.Row() >= numRows {
			return SeatError{seat, SeatErrorCodeOutOfBound, i}
		}

		if seat.Col() < 0 || seat.Col() >= numCols {
			return SeatError{seat, SeatErrorCodeOutOfBound, i}
		}

		idx := seat.Coordinate.AsIndex(numCols)

		if _, ok := m.reservedSeat[idx]; ok {
			return SeatError{seat, SeatErrorCodeSeatTaken, i}
		}

		if _, ok := idxSet[idx]; ok {
			return SeatError{seat, SeatErrorCodeDuplicatedPosition, i}
		}
		idxSet[idx] = struct{}{}

		for occupyIdx, groupID := range m.reservedSeat {
			reservedSeat := Seat{
				GroupID:    groupID,
				Coordinate: m.indexToCoordinate(occupyIdx),
			}
			if !m.isValidDistance(seat, reservedSeat) {
				return SeatError{seat, SeatErrorCodeInvalidDistance, i}
			}
		}
	}

	for _, seat := range seats {
		idx := seat.AsIndex(numCols)
		m.reservedSeat[idx] = seat.GroupID
	}

	return nil
}

func (m *DefaultRoomManager) CancelSeats(_ context.Context, coordinates []Coordinate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	indexes := make([]int64, len(coordinates))
	for i, coord := range coordinates {
		idx := coord.AsIndex(m.cfg.NumCols)
		indexes[i] = idx

		if _, ok := m.reservedSeat[idx]; !ok {
			return SeatError{Seat{Coordinate: coord}, SeatErrorCodeNotReserved, i}
		}
	}

	for _, idx := range indexes {
		delete(m.reservedSeat, idx)
	}

	return nil
}

func (m *DefaultRoomManager) isValidDistance(a, b Seat) bool {
	minDistance := 1
	if a.GroupID != b.GroupID {
		minDistance = m.cfg.MinDistance
	}

	manhattanDist := mathutil.AbsDiff(a.Row(), b.Row()) + mathutil.AbsDiff(a.Col(), b.Col())

	return manhattanDist >= minDistance
}

func (m *DefaultRoomManager) indexToCoordinate(i int64) Coordinate {
	return Coordinate{int(i / int64(m.cfg.NumCols)), int(i % int64(m.cfg.NumCols))}
}

type Seat struct {
	GroupID string
	Coordinate
}

func (s Seat) Row() int {
	return s.Coordinate[0]
}

func (s Seat) Col() int {
	return s.Coordinate[1]
}

type Coordinate [2]int

func (c Coordinate) AsIndex(numCols int) int64 {
	return int64(c[0])*int64(numCols) + int64(c[1])
}
