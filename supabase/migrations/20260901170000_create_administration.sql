alter table public.platform_roles drop constraint platform_roles_role;
alter table public.platform_roles add constraint platform_roles_role check(role in('moderator','administrator'));
alter table public.conversation_reports add column state text not null default 'open';
alter table public.conversation_reports add column resolved_at timestamptz;
alter table public.conversation_reports add column resolved_by_internal_user_id uuid references public.internal_users(id) on delete set null;
alter table public.conversation_reports add constraint conversation_reports_state check(state in('open','resolved'));
create index conversation_reports_open_idx on public.conversation_reports(created_at desc,id) where state='open';
create table public.administration_audit_records(
 id uuid primary key default gen_random_uuid(),
 actor_internal_user_id uuid not null references public.internal_users(id) on delete restrict,
 action text not null,
 target_type text not null,
 target_id uuid not null,
 reason text not null,
 created_at timestamptz not null default timezone('utc',now()),
 constraint administration_audit_action check(action in('hide_review','publish_review','resolve_report')),
 constraint administration_audit_target check(target_type in('review','conversation_report')),
 constraint administration_audit_reason check(char_length(btrim(reason)) between 5 and 500)
);
create index administration_audit_created_idx on public.administration_audit_records(created_at desc,id);
