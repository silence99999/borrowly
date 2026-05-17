package main

import (
	"context"
	"time"
)

type Message struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"sender_id"`
	SenderEmail string    `json:"sender_email"`
	ReceiverID  string    `json:"receiver_id"`
	Content     string    `json:"content"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type Conversation struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	LastMsg   string    `json:"last_message"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (app *application) dbSendMessage(senderID, receiverID, content string) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	msg := &Message{SenderID: senderID, ReceiverID: receiverID, Content: content}
	err := app.db.QueryRowContext(ctx,
		`INSERT INTO chat_messages(sender_id,receiver_id,content) VALUES($1,$2,$3) RETURNING id,is_read,created_at`,
		senderID, receiverID, content,
	).Scan(&msg.ID, &msg.IsRead, &msg.CreatedAt)
	return msg, err
}

func (app *application) dbGetMessages(userID, otherUserID string) ([]*Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT id,sender_id,receiver_id,content,is_read,created_at
		 FROM chat_messages
		 WHERE (sender_id=$1 AND receiver_id=$2) OR (sender_id=$2 AND receiver_id=$1)
		 ORDER BY created_at ASC`,
		userID, otherUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgs := []*Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}

func (app *application) dbGetConversations(userID string) ([]*Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT DISTINCT ON (other_user)
		        CASE WHEN sender_id=$1 THEN receiver_id ELSE sender_id END AS other_user,
		        content, created_at
		 FROM chat_messages
		 WHERE sender_id=$1 OR receiver_id=$1
		 ORDER BY other_user, created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	convs := []*Conversation{}
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.UserID, &c.LastMsg, &c.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, &c)
	}
	return convs, rows.Err()
}
