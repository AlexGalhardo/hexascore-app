package library

import (
	"context"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	vmail "github.com/AlexGalhardo/vukem/mail"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/forms"
	hmail "github.com/AlexGalhardo/hexascore-app/app/mail"
)

// Contact form bounds (ADR 0016 layer 2).
const (
	ContactMinFill = forms.MinFill
	CodeBotCheck   = "bot-check-failed"
)

// ContactTopics are the subjects of the form (legacy lib/contact.ts).
var ContactTopics = []string{"question", "feedback", "bug"}

// Contact stores contact messages and mails them to Inbox.
type Contact struct {
	DB     *database.DB
	Mailer *vmail.Mailer
	// Inbox is CONTACT_INBOX; empty stores only (legacy).
	Inbox string
	// Key signs the form token; derive it with App.DeriveKey.
	Key []byte
	Now func() time.Time
	Log *slog.Logger
}

// ContactInput is a validated form. Name and Email are replaced by the account's when UserID is set.
type ContactInput struct {
	UserID                        *int64
	Name, Email, Subject, Message string
	Honeypot, Token               string
}

// Token returns the signed form token: the render time, so Submit can enforce the minimum fill time.
func (c *Contact) Token() string { return c.forms().Issue("contact") }

func (c *Contact) forms() forms.Token { return forms.Token{Key: c.Key, Now: c.Now} }

// tokenOK checks the signature, the minimum fill time and the expiry.
func (c *Contact) tokenOK(token string) bool { return c.forms().Valid("contact", token) }

// Submit stores the message and queues the mail. A filled honeypot answers success and stores nothing, so a
// bot learns nothing; a bad or too-fast token answers 422 bot-check-failed.
func (c *Contact) Submit(ctx context.Context, in ContactInput) error {
	if in.Honeypot != "" {
		return nil
	}

	if !c.tokenOK(in.Token) {
		return problem(http.StatusUnprocessableEntity, "contact.bot_check", CodeBotCheck, "Please try again.")
	}

	in.Name, in.Email = strings.TrimSpace(in.Name), strings.TrimSpace(in.Email)

	v := errs.NewValidation()
	if in.Name == "" || len([]rune(in.Name)) > 100 {
		v.Append("name", "Enter your name.")
	}

	if a, err := mail.ParseAddress(in.Email); err != nil || a.Address != in.Email {
		v.Append("email", "Enter a valid email.")
	}

	if !v.Empty() {
		return v
	}

	_, err := c.DB.ExecContext(ctx,
		`INSERT INTO contact_messages (id, user_id, name, email, subject, message, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.NewV7().String(), in.UserID, in.Name, in.Email, in.Subject, in.Message, c.Now().UTC())
	if err != nil {
		return err
	}

	if c.Inbox == "" || c.Mailer == nil {
		return nil
	}

	// The message is stored: a failed mail must not make the sender resend it. No PII in the log line.
	if err := c.Mailer.SendLater(ctx, hmail.ContactMessage{
		Inbox: c.Inbox, Name: in.Name, Email: in.Email, Subject: in.Subject, Message: in.Message,
	}); err != nil {
		c.Log.ErrorContext(ctx, "queue contact mail", "err", err)
	}

	return nil
}
