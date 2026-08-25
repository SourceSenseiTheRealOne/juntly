package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reference"
	"github.com/google/uuid"
)

type ReferenceService interface {
	Categories(context.Context, string) ([]reference.Category, error)
	Languages(context.Context, string) ([]reference.Language, error)
	Localities(context.Context, string) ([]reference.Locality, error)
	NearbyLocalities(context.Context, uuid.UUID, int, string) ([]reference.LocalityDistance, error)
}

type referenceKind string

const (
	referenceCategories referenceKind = "categories"
	referenceLanguages  referenceKind = "languages"
	referenceLocalities referenceKind = "localities"
)

type referenceHandler struct {
	service ReferenceService
	kind    referenceKind
}

type categoryResponse struct {
	ID       uuid.UUID  `json:"id"`
	ParentID *uuid.UUID `json:"parentId"`
	Slug     string     `json:"slug"`
	Name     string     `json:"name"`
}

type categoriesResponse struct {
	Categories []categoryResponse `json:"categories"`
}

type languageResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type languagesResponse struct {
	Languages []languageResponse `json:"languages"`
}

type localityResponse struct {
	ID               uuid.UUID `json:"id"`
	Slug             string    `json:"slug"`
	Name             string    `json:"name"`
	ParishName       string    `json:"parishName"`
	MunicipalityName string    `json:"municipalityName"`
	DistrictName     string    `json:"districtName"`
	DistanceMeters   *int      `json:"distanceMeters,omitempty"`
}

type attributionResponse struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type localitiesResponse struct {
	Localities  []localityResponse  `json:"localities"`
	Attribution attributionResponse `json:"attribution"`
}

func NewCategoriesHandler(service ReferenceService) http.Handler {
	return referenceHandler{service: service, kind: referenceCategories}
}

func NewLanguagesHandler(service ReferenceService) http.Handler {
	return referenceHandler{service: service, kind: referenceLanguages}
}

func NewLocalitiesHandler(service ReferenceService) http.Handler {
	return referenceHandler{service: service, kind: referenceLocalities}
}

func (h referenceHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	requestID := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	if h.service == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}
	locale, ok := exactQueryValue(request.URL.Query(), "locale")
	if !ok || !validReferenceLocale(locale) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}

	switch h.kind {
	case referenceCategories:
		if !exactQueryKeys(request.URL.Query(), "locale") {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		values, err := h.service.Categories(request.Context(), locale)
		if err != nil {
			writeReferenceError(w, err, requestID)
			return
		}
		response := categoriesResponse{Categories: make([]categoryResponse, len(values))}
		for index, value := range values {
			response.Categories[index] = categoryResponse{ID: value.ID, ParentID: value.ParentID, Slug: value.Slug, Name: value.Name}
		}
		writeJSON(w, http.StatusOK, response, requestID)
	case referenceLanguages:
		if !exactQueryKeys(request.URL.Query(), "locale") {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		values, err := h.service.Languages(request.Context(), locale)
		if err != nil {
			writeReferenceError(w, err, requestID)
			return
		}
		response := languagesResponse{Languages: make([]languageResponse, len(values))}
		for index, value := range values {
			response.Languages[index] = languageResponse{Code: value.Code, Name: value.Name}
		}
		writeJSON(w, http.StatusOK, response, requestID)
	case referenceLocalities:
		h.serveLocalities(w, request, locale, requestID)
	default:
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
	}
}

func validReferenceLocale(locale string) bool {
	switch locale {
	case "pt-PT", "en", "es":
		return true
	default:
		return false
	}
}

func (h referenceHandler) serveLocalities(w http.ResponseWriter, request *http.Request, locale, requestID string) {
	query := request.URL.Query()
	if !exactQueryKeys(query, "locale", "nearLocalityId", "radiusKm") {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	nearValue, hasNear := optionalExactQueryValue(query, "nearLocalityId")
	radiusValue, hasRadius := optionalExactQueryValue(query, "radiusKm")
	if hasNear != hasRadius {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}

	response := localitiesResponse{Attribution: attributionResponse{Text: reference.AttributionText, URL: reference.AttributionURL}}
	if !hasNear {
		values, err := h.service.Localities(request.Context(), locale)
		if err != nil {
			writeReferenceError(w, err, requestID)
			return
		}
		response.Localities = localityResponses(values)
		writeJSON(w, http.StatusOK, response, requestID)
		return
	}
	origin, err := uuid.Parse(nearValue)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	radius, err := strconv.Atoi(radiusValue)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	values, err := h.service.NearbyLocalities(request.Context(), origin, radius, locale)
	if err != nil {
		writeReferenceError(w, err, requestID)
		return
	}
	response.Localities = make([]localityResponse, len(values))
	for index, value := range values {
		distance := value.DistanceMeters
		response.Localities[index] = localityResponse{ID: value.ID, Slug: value.Slug, Name: value.Name, ParishName: value.ParishName, MunicipalityName: value.MunicipalityName, DistrictName: value.DistrictName, DistanceMeters: &distance}
	}
	writeJSON(w, http.StatusOK, response, requestID)
}

func localityResponses(values []reference.Locality) []localityResponse {
	responses := make([]localityResponse, len(values))
	for index, value := range values {
		responses[index] = localityResponse{ID: value.ID, Slug: value.Slug, Name: value.Name, ParishName: value.ParishName, MunicipalityName: value.MunicipalityName, DistrictName: value.DistrictName}
	}
	return responses
}

func writeReferenceError(w http.ResponseWriter, err error, requestID string) {
	if errors.Is(err, reference.ErrInvalidRequest) || errors.Is(err, reference.ErrNotFound) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
}

func exactQueryKeys(values url.Values, allowed ...string) bool {
	allowlist := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowlist[key] = struct{}{}
	}
	for key := range values {
		if _, ok := allowlist[key]; !ok {
			return false
		}
	}
	return true
}

func exactQueryValue(values url.Values, key string) (string, bool) {
	items, ok := values[key]
	return firstExact(items, ok)
}

func optionalExactQueryValue(values url.Values, key string) (string, bool) {
	items, ok := values[key]
	if !ok {
		return "", false
	}
	return firstExact(items, true)
}

func firstExact(items []string, present bool) (string, bool) {
	returnValue := ""
	if !present || len(items) != 1 || items[0] == "" {
		return returnValue, false
	}
	return items[0], true
}
