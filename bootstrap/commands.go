package bootstrap

import (
	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/console"
)

// Commands lists the application commands; vukem make:command appends here.
func Commands() []vukem.Command {
	return []vukem.Command{
		console.Predeploy(),
		console.CatalogCheck(),
		console.CatalogMerge(),
		console.AdminGrant(),
		console.AdminRevoke(),
		console.UserReset2FA(),
		console.SearchPurge(),
		console.BackupRun(),
		console.BackupRestore(),
		console.BackupDrill(),
		console.ImportLegacy(),
	}
}
