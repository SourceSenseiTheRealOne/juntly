"use client";

import { FormEvent, useEffect, useState } from "react";

type Conversation = {
  id: string;
  listingId: string | null;
  customerId: string;
  providerId: string;
  blocked: boolean;
  createdAt: string;
  updatedAt: string;
};

type Message = {
  id: string;
  conversationId: string;
  senderId: string;
  body: string;
  createdAt: string;
};

export type MessagingCopy = {
  title: string;
  description: string;
  empty: string;
  selectConversation: string;
  messageLabel: string;
  send: string;
  sending: string;
  loading: string;
  error: string;
  conversation: string;
};

export function MessagingInbox({ copy }: { copy: MessagingCopy }) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [selected, setSelected] = useState<string | null>(null);
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let active = true;
    void fetch("/api/v1/me/conversations")
      .then(async (response) => {
        const value: unknown = await response.json();
        if (!response.ok || !validConversations(value)) throw new Error();
        if (active) setConversations(value.conversations);
      })
      .catch(() => {
        if (active) setFailed(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function openConversation(id: string) {
    setFailed(false);
    setSelected(id);
    try {
      const response = await fetch(`/api/v1/me/conversations/${id}/messages`);
      const value: unknown = await response.json();
      if (!response.ok || !validMessages(value)) throw new Error();
      setMessages(value.messages);
    } catch {
      setFailed(true);
    }
  }

  async function send(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const message = body.trim();
    if (!selected || !message || sending) return;
    setFailed(false);
    setSending(true);
    try {
      const response = await fetch(
        `/api/v1/me/conversations/${selected}/messages`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ body: message }),
        },
      );
      const value: unknown = await response.json();
      if (!response.ok || !validMessage(value)) throw new Error();
      setMessages((current) => [...current, value]);
      setBody("");
    } catch {
      setFailed(true);
    } finally {
      setSending(false);
    }
  }

  return (
    <section aria-labelledby="messages-title">
      <h1 id="messages-title" className="text-4xl font-bold tracking-[-0.05em]">
        {copy.title}
      </h1>
      <p className="mt-3 text-lg leading-8 text-muted">{copy.description}</p>

      {loading ? <p className="mt-8 text-muted">{copy.loading}</p> : null}
      {failed ? (
        <p role="alert" className="mt-6 text-earth">
          {copy.error}
        </p>
      ) : null}
      {!loading && conversations.length === 0 ? (
        <p className="market-card mt-8 p-6 text-muted">{copy.empty}</p>
      ) : null}

      {conversations.length > 0 ? (
        <div className="mt-8 grid gap-5 lg:grid-cols-[0.72fr_1.28fr]">
          <nav
            aria-label={copy.selectConversation}
            className="market-card h-fit p-3"
          >
            {conversations.map((conversation, index) => (
              <button
                key={conversation.id}
                type="button"
                className={`min-h-11 w-full rounded-xl px-4 text-left text-sm font-semibold outline-none focus-visible:ring-2 focus-visible:ring-focus ${selected === conversation.id ? "bg-accent-soft text-ink" : "text-muted hover:bg-control"}`}
                onClick={() => void openConversation(conversation.id)}
              >
                {copy.conversation} {index + 1}
              </button>
            ))}
          </nav>

          <div className="market-card min-h-80 p-5">
            {selected ? (
              <>
                <div aria-live="polite" className="grid gap-3">
                  {messages.map((message) => (
                    <p
                      key={message.id}
                      className="max-w-[85%] rounded-xl bg-control px-4 py-3 text-sm leading-6 text-ink"
                    >
                      {message.body}
                    </p>
                  ))}
                </div>
                <form className="mt-6 grid gap-3" onSubmit={send}>
                  <label className="grid gap-2 font-semibold">
                    {copy.messageLabel}
                    <textarea
                      className="market-control min-h-24 py-3"
                      maxLength={4000}
                      value={body}
                      onChange={(event) => setBody(event.target.value)}
                    />
                  </label>
                  <button
                    className="market-button justify-self-start"
                    disabled={sending || !body.trim()}
                    type="submit"
                  >
                    {sending ? copy.sending : copy.send}
                  </button>
                </form>
              </>
            ) : (
              <p className="text-muted">{copy.selectConversation}</p>
            )}
          </div>
        </div>
      ) : null}
    </section>
  );
}

function validConversations(
  value: unknown,
): value is { conversations: Conversation[] } {
  return (
    isRecord(value) &&
    Object.keys(value).length === 1 &&
    Array.isArray(value.conversations) &&
    value.conversations.every(validConversation)
  );
}

function validConversation(value: unknown): value is Conversation {
  return (
    isRecord(value) &&
    uuid(value.id) &&
    (value.listingId === null || uuid(value.listingId)) &&
    uuid(value.customerId) &&
    uuid(value.providerId) &&
    typeof value.blocked === "boolean" &&
    dateTime(value.createdAt) &&
    dateTime(value.updatedAt)
  );
}

function validMessages(value: unknown): value is { messages: Message[] } {
  return (
    isRecord(value) &&
    Object.keys(value).length === 1 &&
    Array.isArray(value.messages) &&
    value.messages.every(validMessage)
  );
}

function validMessage(value: unknown): value is Message {
  return (
    isRecord(value) &&
    uuid(value.id) &&
    uuid(value.conversationId) &&
    uuid(value.senderId) &&
    typeof value.body === "string" &&
    value.body.length > 0 &&
    value.body.length <= 4000 &&
    dateTime(value.createdAt)
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function uuid(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value)
  );
}

function dateTime(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}
