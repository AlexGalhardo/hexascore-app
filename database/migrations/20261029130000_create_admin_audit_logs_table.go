package migrations

import (
	"github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/schema"
)

// The admin panel writes every change it makes to this table.
func m20261029130000CreateAdminAuditLogsTable() schema.Migration {
	return admin.Migration("20261029130000_create_admin_audit_logs_table")
}
