package requests

import "github.com/AlexGalhardo/vukem/z"

// Report is the body of POST /reports (SDD 14 §6.3): a list is reported by id, a user by nickname.
type Report struct {
	TargetType string `json:"targetType" form:"targetType"`
	TargetID   string `json:"targetId"   form:"targetId"`
	Reason     string `json:"reason"     form:"reason"`
	Details    string `json:"details"    form:"details"`
}

// ReportSchema validates Report; the target is resolved by library.Reports.
var ReportSchema = z.Struct[Report](z.Shape{
	"targetType": z.Enum("list", "user", "review", "comment"),
	"targetId":   z.String().Trim().Min(1).Max(64),
	"reason":     z.Enum("spam", "offensive", "harassment", "impersonation", "other"),
	"details":    z.String().Trim().Max(500).Default(""),
})

// ReportResolve is the body of POST /admin/reports/{id}/resolve (SDD 14 §6.4).
type ReportResolve struct {
	Outcome   string `json:"outcome"   form:"outcome"`
	AdminNote string `json:"adminNote" form:"adminNote"`
}

// ReportResolveSchema validates ReportResolve; the outcome must also fit the target (library.Reports.Resolve).
var ReportResolveSchema = z.Struct[ReportResolve](z.Shape{
	"outcome":   z.Enum("dismissed", "list_hidden", "user_banned", "no_action", "content_hidden", "content_deleted"),
	"adminNote": z.String().Trim().Max(1000).Default(""),
})
