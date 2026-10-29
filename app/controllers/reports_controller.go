package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Reports files member reports for one guard: the session guard answers 303 back, the token guard 201 for a new
// report and 200 for a repeat of an open one (SDD 14 §6.3).
type Reports struct {
	svc   *library.Reports
	guard string
	api   bool
}

// NewReports builds the handlers; api selects JSON answers.
func NewReports(svc *library.Reports, guard string, api bool) *Reports {
	return &Reports{svc: svc, guard: guard, api: api}
}

// Create files the report of the signed-in user.
func (h *Reports) Create(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ReportSchema)
	if err != nil {
		return err
	}

	created, err := h.svc.Create(c.Context(), u.ID, library.ReportInput{
		TargetType: in.TargetType, Target: in.TargetID, Reason: in.Reason, Details: in.Details,
	})
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/")
	}

	status := 200
	if created {
		status = 201
	}

	return c.JSON(status, map[string]any{"created": created})
}
