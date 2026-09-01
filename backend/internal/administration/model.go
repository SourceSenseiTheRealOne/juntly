package administration

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalid      = errors.New("administration invalid request")
	ErrUnauthorized = errors.New("administration unauthorized")
	ErrForbidden    = errors.New("administration forbidden")
	ErrNotFound     = errors.New("administration not found")
	ErrUnavailable  = errors.New("administration unavailable")
)

type Metrics struct {
	Users             int `json:"users"`
	Providers         int `json:"providers"`
	ActiveListings    int `json:"activeListings"`
	CompletedBookings int `json:"completedBookings"`
	PublishedReviews  int `json:"publishedReviews"`
	OpenReports       int `json:"openReports"`
}
type ReportItem struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversationId"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"createdAt"`
}
type ReviewItem struct {
	ID        uuid.UUID `json:"id"`
	Rating    int       `json:"rating"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
}
type Queue struct {
	Reports []ReportItem `json:"reports"`
	Reviews []ReviewItem `json:"reviews"`
}
type ModerationAction struct {
	Kind     string    `json:"kind"`
	TargetID uuid.UUID `json:"targetId"`
	Reason   string    `json:"reason"`
}
