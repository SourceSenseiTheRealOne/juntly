package payments

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSQLStorePersistsCheckoutAndIdempotentPaidWebhook(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	database, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	ctx := context.Background()
	customer, provider, booking := uuid.New(), uuid.New(), uuid.New()
	cleanup := func() {
		_, _ = database.ExecContext(ctx, `delete from public.payment_events where payment_order_id in(select id from public.payment_orders where booking_id=$1)`, booking)
		_, _ = database.ExecContext(ctx, `delete from public.stripe_webhook_receipts where provider_object_id='cs_testpayment'`)
		_, _ = database.ExecContext(ctx, `delete from public.payment_orders where booking_id=$1`, booking)
		_, _ = database.ExecContext(ctx, `delete from public.bookings where id=$1`, booking)
		_, _ = database.ExecContext(ctx, `delete from public.provider_payment_accounts where internal_user_id=$1`, provider)
		_, _ = database.ExecContext(ctx, `delete from public.provider_profiles where internal_user_id=$1`, provider)
		_, _ = database.ExecContext(ctx, `delete from public.platform_roles where internal_user_id=$1`, provider)
		_, _ = database.ExecContext(ctx, `delete from public.user_accounts where internal_user_id in($1,$2)`, customer, provider)
		_, _ = database.ExecContext(ctx, `delete from public.internal_users where id in($1,$2)`, customer, provider)
	}
	cleanup()
	defer cleanup()
	for id, subject := range map[uuid.UUID]string{customer: "payment_customer_" + customer.String(), provider: "payment_provider_" + provider.String()} {
		if _, err := database.ExecContext(ctx, `insert into public.internal_users(id,clerk_subject) values($1,$2)`, id, subject); err != nil {
			t.Fatal(err)
		}
		if _, err := database.ExecContext(ctx, `insert into public.user_accounts(internal_user_id,provider_enabled) values($1,$2)`, id, id == provider); err != nil {
			t.Fatal(err)
		}
	}
	var locality uuid.UUID
	if err := database.QueryRowContext(ctx, `select id from public.localities order by id limit 1`).Scan(&locality); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `insert into public.provider_profiles(internal_user_id,display_name,provider_type,bio,primary_locality_id,max_travel_distance_km,remote_services) values($1,'Payment test provider','professional','Synthetic transactional payment test provider.',$2,0,true)`, provider, locality); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `insert into public.bookings(id,customer_internal_user_id,provider_internal_user_id,source_type,idempotency_key,state,scheduled_at,private_location,agreed_price_minor) values($1,$2,$3,'direct','payment-booking-test','confirmed',$4,'Synthetic private test location',12500)`, booking, customer, provider, time.Now().UTC().Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `insert into public.provider_payment_accounts(internal_user_id,stripe_account_id,details_submitted,charges_enabled,payouts_enabled) values($1,'acct_syntheticpayment',true,true,true)`, provider); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `insert into public.platform_roles(id,internal_user_id,role) values($1,$2,'moderator')`, uuid.New(), provider); err != nil {
		t.Fatal(err)
	}
	store := NewSQLStore(database)
	order, account, err := store.PrepareCheckout(ctx, customer, booking, "payment-order-test", 1000)
	if err != nil || order.GrossMinor != 12500 || order.PlatformFeeMinor != 1250 || order.ProviderNetMinor != 11250 || !account.PayoutsEnabled {
		t.Fatalf("prepare order=%#v account=%#v err=%v", order, account, err)
	}
	orderID := uuid.MustParse(order.ID)
	order, err = store.AttachCheckout(ctx, customer, orderID, CheckoutSession{ID: "cs_testpayment", URL: "https://checkout.stripe.test/session"})
	if err != nil || order.State != StateCheckoutCreated {
		t.Fatalf("attach order=%#v err=%v", order, err)
	}
	event := ProviderEvent{ID: "evt_123synthetic", Kind: EventPaid, ProviderObjectID: "cs_testpayment", OrderID: order.ID, PaymentIntentID: "pi_syntheticpayment", InvoiceID: "in_syntheticpayment", OccurredAt: time.Now().UTC()}
	if err := store.ApplyProviderEvent(ctx, event); err != nil {
		t.Fatalf("apply event: %v", err)
	}
	if err := store.ApplyProviderEvent(ctx, event); err != nil {
		t.Fatalf("replay event: %v", err)
	}
	var state string
	var paidEvents, receipts int
	if err := database.QueryRowContext(ctx, `select state from public.payment_orders where id=$1`, orderID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `select count(*) from public.payment_events where payment_order_id=$1 and event_type='paid'`, orderID).Scan(&paidEvents); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `select count(*) from public.stripe_webhook_receipts where stripe_event_id=$1`, event.ID).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if state != "paid" || paidEvents != 1 || receipts != 1 {
		t.Fatalf("state=%s paid_events=%d receipts=%d", state, paidEvents, receipts)
	}
	adminOrders, err := store.ListAdminOrders(ctx, provider)
	if err != nil || len(adminOrders) != 1 || adminOrders[0].ID != order.ID {
		t.Fatalf("admin orders=%#v err=%v", adminOrders, err)
	}
}
