package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/quotations"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxQuotationRequestBytes = 16 * 1024

type QuotationService interface {
	CreateRequest(context.Context, users.VerifiedIdentity, quotations.CreateRequest) (quotations.Request, error)
	ListCustomerRequests(context.Context, users.VerifiedIdentity) ([]quotations.Request, error)
	ListOpportunities(context.Context, users.VerifiedIdentity) ([]quotations.Request, error)
	SubmitProposal(context.Context, users.VerifiedIdentity, uuid.UUID, quotations.SubmitProposal) (quotations.Proposal, error)
	ListProposals(context.Context, users.VerifiedIdentity, uuid.UUID) ([]quotations.Proposal, error)
	AcceptProposal(context.Context, users.VerifiedIdentity, uuid.UUID, uuid.UUID) (quotations.Proposal, error)
}
type quotationHandler struct{ service QuotationService }

func NewQuotationHandler(service QuotationService) http.Handler {
	return quotationHandler{service: service}
}

type quotationRequestBody struct {
	Title            *string    `json:"title"`
	Description      *string    `json:"description"`
	CategoryID       *uuid.UUID `json:"categoryId"`
	LocalityID       *uuid.UUID `json:"localityId"`
	BudgetMinor      *int       `json:"budgetMinor"`
	ProposalDeadline *time.Time `json:"proposalDeadline"`
}
type proposalBody struct {
	PriceMinor       *int       `json:"priceMinor"`
	Message          *string    `json:"message"`
	AvailableAt      *time.Time `json:"availableAt"`
	EstimatedMinutes *int       `json:"estimatedMinutes"`
	ExpiresAt        *time.Time `json:"expiresAt"`
}
type quotationRequestResponse struct {
	ID               uuid.UUID               `json:"id"`
	CustomerID       uuid.UUID               `json:"customerId"`
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	CategoryID       uuid.UUID               `json:"categoryId"`
	LocalityID       uuid.UUID               `json:"localityId"`
	BudgetMinor      *int                    `json:"budgetMinor"`
	ProposalDeadline string                  `json:"proposalDeadline"`
	State            quotations.RequestState `json:"state"`
	CreatedAt        string                  `json:"createdAt"`
	UpdatedAt        string                  `json:"updatedAt"`
}
type proposalResponse struct {
	ID               uuid.UUID                `json:"id"`
	RequestID        uuid.UUID                `json:"requestId"`
	ProviderID       uuid.UUID                `json:"providerId"`
	PriceMinor       int                      `json:"priceMinor"`
	Message          string                   `json:"message"`
	AvailableAt      string                   `json:"availableAt"`
	EstimatedMinutes *int                     `json:"estimatedMinutes"`
	ExpiresAt        *string                  `json:"expiresAt"`
	State            quotations.ProposalState `json:"state"`
	CreatedAt        string                   `json:"createdAt"`
	UpdatedAt        string                   `json:"updatedAt"`
}

func (h quotationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
		return
	}
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "/api/v1/me/quotation-opportunities" {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		values, err := h.service.ListOpportunities(r.Context(), identity)
		if err != nil {
			writeQuotationError(w, err, id)
			return
		}
		writeJSON(w, 200, map[string]any{"requests": requestResponses(values)}, id)
		return
	}
	base := "/api/v1/me/quotation-requests"
	if path == base {
		switch r.Method {
		case http.MethodGet:
			values, err := h.service.ListCustomerRequests(r.Context(), identity)
			if err != nil {
				writeQuotationError(w, err, id)
				return
			}
			writeJSON(w, 200, map[string]any{"requests": requestResponses(values)}, id)
		case http.MethodPost:
			var body quotationRequestBody
			if !decodeQuotation(r.Body, &body) || body.Title == nil || body.Description == nil || body.CategoryID == nil || body.LocalityID == nil || body.ProposalDeadline == nil {
				writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
				return
			}
			value, err := h.service.CreateRequest(r.Context(), identity, quotations.CreateRequest{Title: *body.Title, Description: *body.Description, CategoryID: *body.CategoryID, LocalityID: *body.LocalityID, BudgetMinor: body.BudgetMinor, ProposalDeadline: *body.ProposalDeadline})
			if err != nil {
				writeQuotationError(w, err, id)
				return
			}
			writeJSON(w, 201, requestResponse(value), id)
		default:
			http.Error(w, http.StatusText(405), 405)
		}
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, base+"/"), "/")
	if len(parts) < 2 || parts[1] != "proposals" {
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
		return
	}
	requestID, err := uuid.Parse(parts[0])
	if err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	if len(parts) == 2 {
		switch r.Method {
		case http.MethodGet:
			values, err := h.service.ListProposals(r.Context(), identity, requestID)
			if err != nil {
				writeQuotationError(w, err, id)
				return
			}
			writeJSON(w, 200, map[string]any{"proposals": proposalResponses(values)}, id)
		case http.MethodPost:
			var body proposalBody
			if !decodeQuotation(r.Body, &body) || body.PriceMinor == nil || body.Message == nil || body.AvailableAt == nil {
				writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
				return
			}
			value, err := h.service.SubmitProposal(r.Context(), identity, requestID, quotations.SubmitProposal{PriceMinor: *body.PriceMinor, Message: *body.Message, AvailableAt: *body.AvailableAt, EstimatedMinutes: body.EstimatedMinutes, ExpiresAt: body.ExpiresAt})
			if err != nil {
				writeQuotationError(w, err, id)
				return
			}
			writeJSON(w, 201, proposalJSON(value), id)
		default:
			http.Error(w, http.StatusText(405), 405)
		}
		return
	}
	if len(parts) == 4 && parts[3] == "accept" && r.Method == http.MethodPost {
		proposalID, err := uuid.Parse(parts[2])
		if err != nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		value, err := h.service.AcceptProposal(r.Context(), identity, requestID, proposalID)
		if err != nil {
			writeQuotationError(w, err, id)
			return
		}
		writeJSON(w, 200, proposalJSON(value), id)
		return
	}
	writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
}
func decodeQuotation(body io.Reader, target any) bool {
	d := json.NewDecoder(io.LimitReader(body, maxQuotationRequestBytes+1))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return false
	}
	var extra any
	return errors.Is(d.Decode(&extra), io.EOF)
}
func requestResponse(v quotations.Request) quotationRequestResponse {
	return quotationRequestResponse{ID: v.ID, CustomerID: v.CustomerID, Title: v.Title, Description: v.Description, CategoryID: v.CategoryID, LocalityID: v.LocalityID, BudgetMinor: v.BudgetMinor, ProposalDeadline: v.ProposalDeadline.UTC().Format(time.RFC3339Nano), State: v.State, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func requestResponses(values []quotations.Request) []quotationRequestResponse {
	out := make([]quotationRequestResponse, 0, len(values))
	for _, v := range values {
		out = append(out, requestResponse(v))
	}
	return out
}
func proposalJSON(v quotations.Proposal) proposalResponse {
	var expires *string
	if v.ExpiresAt != nil {
		s := v.ExpiresAt.UTC().Format(time.RFC3339Nano)
		expires = &s
	}
	return proposalResponse{ID: v.ID, RequestID: v.RequestID, ProviderID: v.ProviderID, PriceMinor: v.PriceMinor, Message: v.Message, AvailableAt: v.AvailableAt.UTC().Format(time.RFC3339Nano), EstimatedMinutes: v.EstimatedMinutes, ExpiresAt: expires, State: v.State, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func proposalResponses(values []quotations.Proposal) []proposalResponse {
	out := make([]proposalResponse, 0, len(values))
	for _, v := range values {
		out = append(out, proposalJSON(v))
	}
	return out
}
func writeQuotationError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, quotations.ErrInvalid):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	case errors.Is(err, quotations.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, quotations.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, quotations.ErrNotFound):
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	case errors.Is(err, quotations.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
