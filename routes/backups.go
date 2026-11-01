package routes

import (
	"net/http"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
)

// Backups serves the off-site links that backup:offsite mails (SDD 11 §7.3 step 6): signed, 7 days, monthly
// dumps only, streamed from the S3_* bucket. Without a bucket there is nothing to serve.
func Backups(r *router.Router, app *vukem.App) error {
	disk := controllers.BucketDisk(app)
	if disk == nil {
		return nil
	}

	signer, err := backup.NewOffsiteSigner(app)
	if err != nil {
		return err
	}

	r.Handle(
		"GET "+backup.OffsitePrefix+"/",
		http.StripPrefix(backup.OffsitePrefix, backup.OffsiteHandler(disk, signer)),
	)

	return nil
}
