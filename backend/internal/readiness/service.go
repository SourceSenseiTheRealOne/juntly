package readiness

import "context"

type Pinger interface{ PingContext(context.Context) error }
type Result struct {
	Ready    bool   `json:"ready"`
	Database string `json:"database"`
}
type Service interface{ Check(context.Context) Result }
type service struct{ database Pinger }

func NewService(database Pinger) Service { return service{database: database} }
func (s service) Check(ctx context.Context) Result {
	if s.database == nil || s.database.PingContext(ctx) != nil {
		return Result{Database: "unavailable"}
	}
	return Result{Ready: true, Database: "ready"}
}
