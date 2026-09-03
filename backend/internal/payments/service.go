package payments

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

var paymentIdempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{8,128}$`)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type ModeratorAuthorizer interface {
	RequireModerator(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}

type Store interface {
	PrepareCheckout(context.Context, uuid.UUID, uuid.UUID, string, int) (Order, ProviderAccount, error)
	AttachCheckout(context.Context, uuid.UUID, uuid.UUID, CheckoutSession) (Order, error)
	ListOrders(context.Context, uuid.UUID) ([]Order, error)
	ListAdminOrders(context.Context, uuid.UUID) ([]Order, error)
	GetProviderAccount(context.Context, uuid.UUID) (ProviderAccount, error)
	SaveProviderAccount(context.Context, uuid.UUID, ConnectedAccount) (ProviderAccount, error)
	ApplyProviderEvent(context.Context, ProviderEvent) error
	PrepareRefund(context.Context, uuid.UUID, uuid.UUID) (Order, error)
	AttachRefund(context.Context, uuid.UUID, uuid.UUID, RefundResult) (Order, error)
}

type Gateway interface {
	CreateCheckout(context.Context, CheckoutRequest) (CheckoutSession, error)
	CreateConnectedAccount(context.Context, string) (ConnectedAccount, error)
	CreateAccountLink(context.Context, string, string) (AccountLink, error)
	GetConnectedAccount(context.Context, string) (ConnectedAccount, error)
	CreateRefund(context.Context, string, string) (RefundResult, error)
	VerifyWebhook([]byte, string) (ProviderEvent, error)
}

type CheckoutResult struct {
	Order Order  `json:"order"`
	URL   string `json:"url"`
}

type PayoutOnboardingResult struct {
	Account ProviderAccount `json:"account"`
	URL     string          `json:"url"`
}

type Service interface {
	BeginCheckout(context.Context, users.VerifiedIdentity, uuid.UUID, string, string) (CheckoutResult, error)
	ListOrders(context.Context, users.VerifiedIdentity) ([]Order, error)
	ListAdminOrders(context.Context, users.VerifiedIdentity) ([]Order, error)
	BeginPayoutOnboarding(context.Context, users.VerifiedIdentity, string) (PayoutOnboardingResult, error)
	PayoutStatus(context.Context, users.VerifiedIdentity) (ProviderAccount, error)
	HandleWebhook(context.Context, []byte, string) error
	Refund(context.Context, users.VerifiedIdentity, uuid.UUID, string) (Order, error)
}

type service struct {
	identities IdentityReconciler
	moderators ModeratorAuthorizer
	store      Store
	gateway    Gateway
	feeBPS     int
}

func NewService(identities IdentityReconciler, moderators ModeratorAuthorizer, store Store, gateway Gateway, feeBPS int) Service {
	return service{identities: identities, moderators: moderators, store: store, gateway: gateway, feeBPS: feeBPS}
}

func (s service) BeginCheckout(ctx context.Context, identity users.VerifiedIdentity, bookingID uuid.UUID, idempotencyKey, locale string) (CheckoutResult, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if bookingID == uuid.Nil || !paymentIdempotencyPattern.MatchString(idempotencyKey) || s.feeBPS < 0 || s.feeBPS >= 10_000 {
		return CheckoutResult{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return CheckoutResult{}, err
	}
	if s.gateway == nil {
		return CheckoutResult{}, ErrUnavailable
	}
	order, account, err := s.store.PrepareCheckout(ctx, actor.ID, bookingID, idempotencyKey, s.feeBPS)
	if err != nil {
		return CheckoutResult{}, normalize(err)
	}
	if !account.DetailsSubmitted || !account.ChargesEnabled || !account.PayoutsEnabled {
		return CheckoutResult{}, ErrForbidden
	}
	session, err := s.gateway.CreateCheckout(ctx, CheckoutRequest{OrderID: order.ID, BookingID: order.BookingID, ConnectedAccountID: account.StripeAccountID, GrossMinor: order.GrossMinor, FeeMinor: order.PlatformFeeMinor, Locale: locale, IdempotencyKey: "checkout-" + order.ID})
	if err != nil {
		return CheckoutResult{}, normalize(err)
	}
	orderID, err := uuid.Parse(order.ID)
	if err != nil {
		return CheckoutResult{}, ErrUnavailable
	}
	attached, err := s.store.AttachCheckout(ctx, actor.ID, orderID, session)
	if err != nil {
		return CheckoutResult{}, normalize(err)
	}
	return CheckoutResult{Order: attached, URL: session.URL}, nil
}

func (s service) ListOrders(ctx context.Context, identity users.VerifiedIdentity) ([]Order, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	orders, err := s.store.ListOrders(ctx, actor.ID)
	return orders, normalize(err)
}

func (s service) ListAdminOrders(ctx context.Context, identity users.VerifiedIdentity) ([]Order, error) {
	if s.moderators == nil || s.store == nil {
		return nil, ErrUnavailable
	}
	moderator, err := s.moderators.RequireModerator(ctx, identity)
	if err != nil {
		return nil, normalize(err)
	}
	orders, err := s.store.ListAdminOrders(ctx, moderator.ID)
	return orders, normalize(err)
}

func (s service) BeginPayoutOnboarding(ctx context.Context, identity users.VerifiedIdentity, locale string) (PayoutOnboardingResult, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return PayoutOnboardingResult{}, err
	}
	if s.gateway == nil {
		return PayoutOnboardingResult{}, ErrUnavailable
	}
	account, err := s.store.GetProviderAccount(ctx, actor.ID)
	if errors.Is(err, ErrNotFound) {
		created, createErr := s.gateway.CreateConnectedAccount(ctx, "connect-"+actor.ID.String())
		if createErr != nil {
			return PayoutOnboardingResult{}, normalize(createErr)
		}
		account, err = s.store.SaveProviderAccount(ctx, actor.ID, created)
	} else if err == nil {
		refreshed, refreshErr := s.gateway.GetConnectedAccount(ctx, account.StripeAccountID)
		if refreshErr != nil {
			return PayoutOnboardingResult{}, normalize(refreshErr)
		}
		account, err = s.store.SaveProviderAccount(ctx, actor.ID, refreshed)
	}
	if err != nil {
		return PayoutOnboardingResult{}, normalize(err)
	}
	link, err := s.gateway.CreateAccountLink(ctx, account.StripeAccountID, supportedLocale(locale))
	if err != nil {
		return PayoutOnboardingResult{}, normalize(err)
	}
	return PayoutOnboardingResult{Account: account, URL: link.URL}, nil
}

func (s service) PayoutStatus(ctx context.Context, identity users.VerifiedIdentity) (ProviderAccount, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return ProviderAccount{}, err
	}
	account, err := s.store.GetProviderAccount(ctx, actor.ID)
	if err != nil {
		return ProviderAccount{}, normalize(err)
	}
	if s.gateway == nil {
		return ProviderAccount{}, ErrUnavailable
	}
	refreshed, err := s.gateway.GetConnectedAccount(ctx, account.StripeAccountID)
	if err != nil {
		return ProviderAccount{}, normalize(err)
	}
	account, err = s.store.SaveProviderAccount(ctx, actor.ID, refreshed)
	return account, normalize(err)
}

func (s service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	if s.gateway == nil || s.store == nil {
		return ErrUnavailable
	}
	event, err := s.gateway.VerifyWebhook(payload, signature)
	if err != nil {
		return normalize(err)
	}
	return normalize(s.store.ApplyProviderEvent(ctx, event))
}

func (s service) Refund(ctx context.Context, identity users.VerifiedIdentity, orderID uuid.UUID, idempotencyKey string) (Order, error) {
	if orderID == uuid.Nil || !paymentIdempotencyPattern.MatchString(idempotencyKey) || s.moderators == nil || s.gateway == nil || s.store == nil {
		return Order{}, ErrInvalid
	}
	moderator, err := s.moderators.RequireModerator(ctx, identity)
	if err != nil {
		return Order{}, normalize(err)
	}
	order, err := s.store.PrepareRefund(ctx, moderator.ID, orderID)
	if err != nil {
		return Order{}, normalize(err)
	}
	if order.State == StateRefundPending {
		return order, nil
	}
	refund, err := s.gateway.CreateRefund(ctx, order.PaymentIntentID, "refund-"+order.ID)
	if err != nil {
		return Order{}, normalize(err)
	}
	order, err = s.store.AttachRefund(ctx, moderator.ID, orderID, refund)
	return order, normalize(err)
}

func (s service) actor(ctx context.Context, identity users.VerifiedIdentity) (users.InternalUser, error) {
	if s.identities == nil || s.store == nil {
		return users.InternalUser{}, ErrUnavailable
	}
	actor, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil || actor.ID == uuid.Nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return users.InternalUser{}, ErrUnauthorized
		}
		return users.InternalUser{}, ErrUnavailable
	}
	return actor, nil
}

func supportedLocale(locale string) string {
	if locale == "en" || locale == "es" || locale == "pt-PT" {
		return locale
	}
	return "pt-PT"
}

func normalize(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{ErrInvalid, ErrUnauthorized, ErrForbidden, ErrNotFound, ErrConflict} {
		if errors.Is(err, known) {
			return known
		}
	}
	return ErrUnavailable
}
