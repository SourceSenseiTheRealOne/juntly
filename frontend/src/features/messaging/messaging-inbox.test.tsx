import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { MessagingInbox } from "./messaging-inbox";

const copy = {
  title: "Mensagens",
  description: "Converse em privado.",
  empty: "Ainda não tem conversas.",
  selectConversation: "Selecionar conversa",
  messageLabel: "Mensagem",
  send: "Enviar",
  sending: "A enviar...",
  loading: "A carregar...",
  error: "Não foi possível carregar.",
  conversation: "Conversa",
};

afterEach(() => vi.restoreAllMocks());

describe("MessagingInbox", () => {
  it("loads participant conversations and sends a private message", async () => {
    const conversationId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        Response.json({
          conversations: [{ id: conversationId, listingId: null, customerId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", providerId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc", blocked: false, createdAt: "2026-09-01T12:00:00Z", updatedAt: "2026-09-01T12:00:00Z" }],
        }),
      )
      .mockResolvedValueOnce(Response.json({ messages: [] }))
      .mockResolvedValueOnce(
        Response.json({ id: "dddddddd-dddd-4ddd-8ddd-dddddddddddd", conversationId, senderId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", body: "Boa tarde", createdAt: "2026-09-01T12:01:00Z" }, { status: 201 }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<MessagingInbox copy={copy} />);
    const conversation = await screen.findByRole("button", { name: /Conversa 1/ });
    fireEvent.click(conversation);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    fireEvent.change(screen.getByLabelText(copy.messageLabel), { target: { value: "Boa tarde" } });
    fireEvent.click(screen.getByRole("button", { name: copy.send }));

    expect(await screen.findByText("Boa tarde")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenLastCalledWith(
      `/api/v1/me/conversations/${conversationId}/messages`,
      expect.objectContaining({ method: "POST" }),
    );
  });
});
