import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { SectionState } from "../components/SectionState";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";
import { formatDateTime } from "../lib/utils";
import type { ChatMessage, Conversation } from "../types/api";

export function ChatPage() {
  const { userId } = useParams<{ userId?: string }>();
  const { user } = useAuth();
  const navigate = useNavigate();

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [content, setContent] = useState("");
  const [error, setError] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);

  async function loadConversations() {
    try {
      setConversations(await api.getConversations());
    } catch {
      // empty inbox is fine
    }
  }

  async function loadMessages(targetId: string) {
    try {
      const msgs = await api.getMessages(targetId);
      setMessages(msgs);
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load messages.");
    }
  }

  useEffect(() => {
    void loadConversations();
  }, []);

  useEffect(() => {
    if (userId) void loadMessages(userId);
    else setMessages([]);
  }, [userId]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSend(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!userId || !content.trim()) return;
    try {
      await api.sendMessage(userId, content.trim());
      setContent("");
      await loadMessages(userId);
      await loadConversations();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to send message.");
    }
  }

  return (
    <div className="stack-page">
      <section className="panel-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Messages</p>
            <h1>Chat</h1>
          </div>
        </div>

        <div className="chat-layout">
          <aside className="chat-sidebar">
            {conversations.length === 0 ? (
              <SectionState title="No conversations" description="Start one from an item's detail page." />
            ) : null}
            {conversations.map((conv) => (
              <button
                key={conv.user_id}
                className={`chat-conv-item ${userId === conv.user_id ? "active" : ""}`}
                onClick={() => navigate(`/chat/${conv.user_id}`)}
                type="button"
              >
                <strong className="chat-conv-id">{conv.email || conv.user_id.slice(0, 8) + "…"}</strong>
                <span className="chat-conv-last">{conv.last_message}</span>
              </button>
            ))}
          </aside>

          <section className="chat-main">
            {!userId ? (
              <SectionState title="Select a conversation" description="Pick one from the left or open a chat from an item page." />
            ) : (
              <>
                <div className="chat-messages">
                  {messages.length === 0 ? (
                    <SectionState title="No messages yet" description="Send the first message below." />
                  ) : null}
                  {messages.map((msg) => {
                    const mine = msg.sender_id === user?.id;
                    return (
                      <div key={msg.id} className={`chat-bubble ${mine ? "mine" : "theirs"}`}>
                        <p>{msg.content}</p>
                        <small>{formatDateTime(msg.created_at)}</small>
                      </div>
                    );
                  })}
                  <div ref={bottomRef} />
                </div>

                {error ? <p className="form-error">{error}</p> : null}

                <form className="chat-input-row" onSubmit={handleSend}>
                  <input
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    placeholder="Type a message…"
                    required
                  />
                  <button className="button" type="submit">
                    Send
                  </button>
                </form>
              </>
            )}
          </section>
        </div>
      </section>
    </div>
  );
}
