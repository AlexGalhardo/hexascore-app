package admin

import (
	vadmin "github.com/AlexGalhardo/vukem/admin"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ItemsResource lists and shows catalogue items. It has no writable field: merges and affiliate links go through
// the routes of CatalogHandlers (VF-008, no panel row actions).
func ItemsResource() *vadmin.Resource[models.Item] {
	return vadmin.NewResource[models.Item](noInputSchema).
		Slug("items").
		Search("slug", "wikidata_qid").
		Filter("type", "game", "movie", "series", "book").
		Filter("visibility", "public", "hidden").
		List("id", "type", "slug", "wikidata_qid", "visibility", "updated_at")
}
