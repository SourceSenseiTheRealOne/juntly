package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/discovery"
	"github.com/google/uuid"
)

type PublicDiscoveryService interface {
	Search(context.Context, discovery.Request) ([]discovery.Listing, error)
	Get(context.Context, string, string) (*discovery.Listing, error)
}

type publicDiscoveryHandler struct{ service PublicDiscoveryService }
type publicListingHandler struct{ service PublicDiscoveryService }

type publicListingResponse struct {
	ID                  uuid.UUID `json:"id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	CategoryID          uuid.UUID `json:"categoryId"`
	CategorySlug        string    `json:"categorySlug"`
	CategoryName        string    `json:"categoryName"`
	PrimaryLocalityID   uuid.UUID `json:"primaryLocalityId"`
	LocalitySlug        string    `json:"localitySlug"`
	LocalityName        string    `json:"localityName"`
	PriceType           string    `json:"priceType"`
	PriceMinor          *int      `json:"priceMinor"`
	Currency            string    `json:"currency"`
	TravelsToCustomer   bool      `json:"travelsToCustomer"`
	ReceivesCustomer    bool      `json:"receivesCustomer"`
	RemoteServices      bool      `json:"remoteServices"`
	ProviderDisplayName string    `json:"providerDisplayName"`
	ProviderType        string    `json:"providerType"`
	UpdatedAt           string    `json:"updatedAt"`
}

type publicListingsResponse struct {
	Listings []publicListingResponse `json:"listings"`
}

func NewPublicDiscoveryHandler(service PublicDiscoveryService) http.Handler {
	return publicDiscoveryHandler{service: service}
}

func NewPublicListingHandler(service PublicDiscoveryService) http.Handler {
	return publicListingHandler{service: service}
}

func (h publicDiscoveryHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	id := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	if request.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if h.service == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	query, ok := publicDiscoveryRequest(request.URL.Query())
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	values, err := h.service.Search(request.Context(), query)
	if err != nil {
		writePublicDiscoveryError(w, err, id)
		return
	}
	response := publicListingsResponse{Listings: make([]publicListingResponse, len(values))}
	for index, value := range values {
		response.Listings[index] = publicListingResponseFrom(value)
	}
	writeJSON(w, http.StatusOK, response, id)
}

func (h publicListingHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	id := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	if request.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if h.service == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	if !exactQueryKeys(request.URL.Query(), "locale") {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	locale, ok := exactQueryValue(request.URL.Query(), "locale")
	if !ok || !validReferenceLocale(locale) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	rawID := strings.TrimPrefix(request.URL.Path, "/api/v1/public/listings/")
	if rawID == request.URL.Path || strings.Contains(rawID, "/") {
		writePublicNotFound(w, id)
		return
	}
	value, err := h.service.Get(request.Context(), rawID, locale)
	if errors.Is(err, discovery.ErrNotFound) || value == nil {
		writePublicNotFound(w, id)
		return
	}
	if err != nil {
		writePublicDiscoveryError(w, err, id)
		return
	}
	writeJSON(w, http.StatusOK, publicListingResponseFrom(*value), id)
}

func publicDiscoveryRequest(values url.Values) (discovery.Request, bool) {
	if !exactQueryKeys(values, "locale", "categoryId", "q", "nearLocalityId", "radiusKm", "priceType", "serviceMode") {
		return discovery.Request{}, false
	}
	locale, ok := exactQueryValue(values, "locale")
	if !ok || !validReferenceLocale(locale) {
		return discovery.Request{}, false
	}
	request := discovery.Request{Locale: locale}
	var present bool
	if request.CategoryID, present = optionalUUIDQuery(values, "categoryId"); !present {
		return discovery.Request{}, false
	}
	if request.Query, present = optionalExactQueryValue(values, "q"); !present && values.Has("q") {
		return discovery.Request{}, false
	}
	var near string
	near, present = optionalExactQueryValue(values, "nearLocalityId")
	if !present && values.Has("nearLocalityId") {
		return discovery.Request{}, false
	}
	radius, hasRadius := optionalExactQueryValue(values, "radiusKm")
	if !hasRadius && values.Has("radiusKm") {
		return discovery.Request{}, false
	}
	if present != hasRadius {
		return discovery.Request{}, false
	}
	if present {
		parsed, err := uuid.Parse(near)
		if err != nil {
			return discovery.Request{}, false
		}
		request.NearLocalityID = parsed
		request.RadiusKM, err = strconv.Atoi(radius)
		if err != nil {
			return discovery.Request{}, false
		}
	}
	if request.PriceType, present = optionalPriceTypeQuery(values); !present {
		return discovery.Request{}, false
	}
	if request.ServiceMode, present = optionalServiceModeQuery(values); !present {
		return discovery.Request{}, false
	}
	return request, true
}

func optionalUUIDQuery(values url.Values, key string) (uuid.UUID, bool) {
	value, present := optionalExactQueryValue(values, key)
	if !present {
		return uuid.Nil, !values.Has(key)
	}
	parsed, err := uuid.Parse(value)
	return parsed, err == nil
}

func optionalPriceTypeQuery(values url.Values) (discovery.PriceType, bool) {
	value, present := optionalExactQueryValue(values, "priceType")
	if !present {
		return "", !values.Has("priceType")
	}
	candidate := discovery.PriceType(value)
	switch candidate {
	case discovery.PriceTypeFixed, discovery.PriceTypeHourly, discovery.PriceTypeDaily, discovery.PriceTypeQuote, discovery.PriceTypeNegotiable:
		return candidate, true
	default:
		return "", false
	}
}

func optionalServiceModeQuery(values url.Values) (discovery.ServiceMode, bool) {
	value, present := optionalExactQueryValue(values, "serviceMode")
	if !present {
		return "", !values.Has("serviceMode")
	}
	candidate := discovery.ServiceMode(value)
	switch candidate {
	case discovery.ServiceModeTravelsToCustomer, discovery.ServiceModeReceivesCustomer, discovery.ServiceModeRemoteServices:
		return candidate, true
	default:
		return "", false
	}
}

func publicListingResponseFrom(value discovery.Listing) publicListingResponse {
	return publicListingResponse{
		ID: value.ID, Title: value.Title, Description: value.Description,
		CategoryID: value.CategoryID, CategorySlug: value.CategorySlug, CategoryName: value.CategoryName,
		PrimaryLocalityID: value.PrimaryLocalityID, LocalitySlug: value.LocalitySlug, LocalityName: value.LocalityName,
		PriceType: string(value.PriceType), PriceMinor: value.PriceMinor, Currency: value.Currency,
		TravelsToCustomer: value.TravelsToCustomer, ReceivesCustomer: value.ReceivesCustomer, RemoteServices: value.RemoteServices,
		ProviderDisplayName: value.ProviderDisplayName, ProviderType: value.ProviderType,
		UpdatedAt: value.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func writePublicDiscoveryError(w http.ResponseWriter, err error, id string) {
	if errors.Is(err, discovery.ErrInvalidRequest) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", id)
}

func writePublicNotFound(w http.ResponseWriter, id string) {
	writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Not found", id)
}
