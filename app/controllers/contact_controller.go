package controllers

import (
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Contact serves the public contact form (SDD 14 §6.2, §7.3).
type Contact struct {
	svc  *library.Contact
	user func(c *web.Ctx) (models.User, error)
}

// NewContact builds the handlers; user returns the signed-in user, if any.
func NewContact(svc *library.Contact, user func(c *web.Ctx) (models.User, error)) *Contact {
	return &Contact{svc: svc, user: user}
}

// Index renders the form with a fresh token.
func (h *Contact) Index(c *web.Ctx) error {
	return inertia.Render(c, "Contact/Index", inertia.Props{
		"topics": library.ContactTopics, "formToken": h.svc.Token(),
	})
}

// Send stores the message. A signed-in user cannot speak for someone else: name and email come from the account.
func (h *Contact) Send(c *web.Ctx) error {
	in, err := web.Parse(c, requests.ContactSchema)
	if err != nil {
		return err
	}

	ci := library.ContactInput{
		Name: in.Name, Email: in.Email, Subject: in.Subject, Message: in.Message,
		Honeypot: in.Website, Token: in.Token,
	}
	if u, err := h.user(c); err == nil {
		ci.UserID, ci.Name, ci.Email = &u.ID, u.Name, u.Email
	}

	if err := h.svc.Submit(c.Context(), ci); err != nil {
		return err
	}

	return c.Back("/contact")
}
