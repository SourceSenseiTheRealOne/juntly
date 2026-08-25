package health

import "time"

const ServiceName = "juntly-api"

type Clock func() time.Time

type Service struct {
	version string
	now     Clock
}

type Result struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	CheckedAt time.Time `json:"checkedAt"`
	RequestID string    `json:"requestId"`
}

func NewService(version string, now Clock) Service {
	if version == "" {
		version = "dev"
	}
	if now == nil {
		now = time.Now
	}

	return Service{
		version: version,
		now:     now,
	}
}

func (s Service) Check(requestID string) Result {
	return Result{
		Status:    "ok",
		Service:   ServiceName,
		Version:   s.version,
		CheckedAt: s.now().UTC(),
		RequestID: requestID,
	}
}
