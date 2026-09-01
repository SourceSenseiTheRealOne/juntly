package messaging

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(database *sql.DB) Store { return sqlStore{database: database} }

func (s sqlStore) Start(ctx context.Context, actorID, listingID uuid.UUID) (Conversation, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Conversation{}, err
	}
	defer tx.Rollback()
	var providerID uuid.UUID
	if err := tx.QueryRowContext(ctx, `select internal_user_id from public.listings where id=$1 and state='active'`, listingID).Scan(&providerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, ErrNotFound
		}
		return Conversation{}, err
	}
	if providerID == actorID {
		return Conversation{}, ErrForbidden
	}
	var value Conversation
	err = tx.QueryRowContext(ctx, `insert into public.conversations (listing_id,customer_internal_user_id,provider_internal_user_id) values ($1,$2,$3)
		on conflict (listing_id,customer_internal_user_id,provider_internal_user_id) do update set updated_at=public.conversations.updated_at
		returning id,listing_id,customer_internal_user_id,provider_internal_user_id,blocked_by_internal_user_id,created_at,updated_at`, listingID, actorID, providerID).
		Scan(&value.ID, &value.ListingID, &value.CustomerID, &value.ProviderID, &value.BlockedBy, &value.CreatedAt, &value.UpdatedAt)
	if err != nil {
		return Conversation{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notifications (recipient_internal_user_id,kind,resource_id,in_app_visible)
		select $1,'conversation_started',$2,coalesce((select in_app_enabled from public.notification_preferences where internal_user_id=$1),true)
		where coalesce((select in_app_enabled or email_enabled from public.notification_preferences where internal_user_id=$1),true)
		and not exists (select 1 from public.notifications where recipient_internal_user_id=$1 and kind='conversation_started' and resource_id=$2)`, providerID, value.ID); err != nil {
		return Conversation{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notification_email_outbox (notification_id,recipient_internal_user_id)
		select id,recipient_internal_user_id from public.notifications where recipient_internal_user_id=$1 and kind='conversation_started' and resource_id=$2
		and coalesce((select email_enabled from public.notification_preferences where internal_user_id=$1),true)
		on conflict (notification_id) do nothing`, providerID, value.ID); err != nil {
		return Conversation{}, err
	}
	if err = tx.Commit(); err != nil {
		return Conversation{}, err
	}
	return value, nil
}
func (s sqlStore) List(ctx context.Context, actorID uuid.UUID) ([]Conversation, error) {
	rows, err := s.database.QueryContext(ctx, `select id,listing_id,customer_internal_user_id,provider_internal_user_id,blocked_by_internal_user_id,created_at,updated_at from public.conversations where customer_internal_user_id=$1 or provider_internal_user_id=$1 order by updated_at desc,id`, actorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Conversation{}
	for rows.Next() {
		var value Conversation
		if err := rows.Scan(&value.ID, &value.ListingID, &value.CustomerID, &value.ProviderID, &value.BlockedBy, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func (s sqlStore) ListMessages(ctx context.Context, actorID, conversationID uuid.UUID) ([]Message, error) {
	rows, err := s.database.QueryContext(ctx, `select m.id,m.conversation_id,m.sender_internal_user_id,m.body,m.created_at from public.conversation_messages m join public.conversations c on c.id=m.conversation_id where c.id=$1 and ($2=c.customer_internal_user_id or $2=c.provider_internal_user_id) order by m.created_at,m.id`, conversationID, actorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Message{}
	for rows.Next() {
		var value Message
		if err := rows.Scan(&value.ID, &value.ConversationID, &value.SenderID, &value.Body, &value.CreatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func (s sqlStore) Send(ctx context.Context, actorID, conversationID uuid.UUID, body string) (Message, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, err
	}
	defer tx.Rollback()
	var customerID, providerID uuid.UUID
	var blockedBy *uuid.UUID
	if err = tx.QueryRowContext(ctx, `select customer_internal_user_id,provider_internal_user_id,blocked_by_internal_user_id from public.conversations where id=$1 for update`, conversationID).Scan(&customerID, &providerID, &blockedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, ErrNotFound
		}
		return Message{}, err
	}
	if actorID != customerID && actorID != providerID {
		return Message{}, ErrForbidden
	}
	if blockedBy != nil {
		return Message{}, ErrForbidden
	}
	var value Message
	err = tx.QueryRowContext(ctx, `insert into public.conversation_messages (conversation_id,sender_internal_user_id,body) values ($1,$2,$3) returning id,conversation_id,sender_internal_user_id,body,created_at`, conversationID, actorID, body).Scan(&value.ID, &value.ConversationID, &value.SenderID, &value.Body, &value.CreatedAt)
	if err != nil {
		return Message{}, err
	}
	recipient := customerID
	if actorID == customerID {
		recipient = providerID
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notifications (recipient_internal_user_id,kind,resource_id,in_app_visible)
		select $1,'message_received',$2,coalesce((select in_app_enabled from public.notification_preferences where internal_user_id=$1),true)
		where coalesce((select in_app_enabled or email_enabled from public.notification_preferences where internal_user_id=$1),true)`, recipient, value.ID); err != nil {
		return Message{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notification_email_outbox (notification_id,recipient_internal_user_id)
		select id,recipient_internal_user_id from public.notifications where recipient_internal_user_id=$1 and kind='message_received' and resource_id=$2
		and coalesce((select email_enabled from public.notification_preferences where internal_user_id=$1),true)`, recipient, value.ID); err != nil {
		return Message{}, err
	}
	if _, err = tx.ExecContext(ctx, `update public.conversations set updated_at=timezone('utc',now()) where id=$1`, conversationID); err != nil {
		return Message{}, err
	}
	if err = tx.Commit(); err != nil {
		return Message{}, err
	}
	return value, nil
}
func (s sqlStore) SetBlocked(ctx context.Context, actorID, conversationID uuid.UUID, blocked bool) error {
	var result sql.Result
	var err error
	if blocked {
		result, err = s.database.ExecContext(ctx, `update public.conversations set blocked_by_internal_user_id=$1,updated_at=timezone('utc',now()) where id=$2 and ($1=customer_internal_user_id or $1=provider_internal_user_id)`, actorID, conversationID)
	} else {
		result, err = s.database.ExecContext(ctx, `update public.conversations set blocked_by_internal_user_id=null,updated_at=timezone('utc',now()) where id=$2 and blocked_by_internal_user_id=$1`, actorID, conversationID)
	}
	return affected(result, err)
}
func (s sqlStore) Report(ctx context.Context, actorID, conversationID uuid.UUID, messageID *uuid.UUID, reason string) error {
	result, err := s.database.ExecContext(ctx, `insert into public.conversation_reports (conversation_id,message_id,reporter_internal_user_id,reason) select $1,$2,$3,$4 where exists (select 1 from public.conversations where id=$1 and ($3=customer_internal_user_id or $3=provider_internal_user_id)) and ($2::uuid is null or exists(select 1 from public.conversation_messages where id=$2 and conversation_id=$1))`, conversationID, messageID, actorID, reason)
	return affected(result, err)
}
func (s sqlStore) Preferences(ctx context.Context, actorID uuid.UUID) (NotificationPreferences, error) {
	var v NotificationPreferences
	err := s.database.QueryRowContext(ctx, `select coalesce((select in_app_enabled from public.notification_preferences where internal_user_id=$1),true),coalesce((select email_enabled from public.notification_preferences where internal_user_id=$1),true)`, actorID).Scan(&v.InAppEnabled, &v.EmailEnabled)
	return v, err
}
func (s sqlStore) ReplacePreferences(ctx context.Context, actorID uuid.UUID, v NotificationPreferences) (NotificationPreferences, error) {
	err := s.database.QueryRowContext(ctx, `insert into public.notification_preferences (internal_user_id,in_app_enabled,email_enabled) values($1,$2,$3) on conflict(internal_user_id) do update set in_app_enabled=excluded.in_app_enabled,email_enabled=excluded.email_enabled,updated_at=timezone('utc',now()) returning in_app_enabled,email_enabled`, actorID, v.InAppEnabled, v.EmailEnabled).Scan(&v.InAppEnabled, &v.EmailEnabled)
	return v, err
}
func (s sqlStore) Notifications(ctx context.Context, actorID uuid.UUID) ([]Notification, error) {
	rows, err := s.database.QueryContext(ctx, `select id,kind,read_at,created_at from public.notifications where recipient_internal_user_id=$1 and in_app_visible order by created_at desc,id limit 100`, actorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Notification{}
	for rows.Next() {
		var v Notification
		if err := rows.Scan(&v.ID, &v.Kind, &v.ReadAt, &v.CreatedAt); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func (s sqlStore) MarkNotificationRead(ctx context.Context, actorID, notificationID uuid.UUID) error {
	result, err := s.database.ExecContext(ctx, `update public.notifications set read_at=coalesce(read_at,timezone('utc',now())) where id=$1 and recipient_internal_user_id=$2`, notificationID, actorID)
	return affected(result, err)
}
func affected(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrForbidden
	}
	return nil
}

var _ = time.Time{}
