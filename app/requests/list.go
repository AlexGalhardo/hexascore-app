package requests

import "github.com/AlexGalhardo/vukem/z"

// ListCreate is the new-list form (SDD 14 §6.3).
type ListCreate struct {
	Type        string `json:"type"        form:"type"`
	Title       string `json:"title"       form:"title"`
	Description string `json:"description" form:"description"`
	IsPublic    bool   `json:"isPublic"    form:"isPublic"`
	FirstItemID string `json:"firstItemId" form:"firstItemId"`
}

// ListCreateSchema validates ListCreate; list limits and the item's type are checked by library.Lists.
var ListCreateSchema = z.Struct[ListCreate](z.Shape{
	"type":        z.Enum("game", "movie", "series", "book"),
	"title":       z.String().Trim().Min(1).Max(80),
	"description": z.String().Trim().Max(500).Default(""),
	"isPublic":    z.Bool().Default(true),
	"firstItemId": z.String().UUID().Default(""),
})

// ListUpdate patches a list; absent fields stay as they are.
type ListUpdate struct {
	Title       *string `json:"title"       form:"title"`
	Description *string `json:"description" form:"description"`
	IsPublic    *bool   `json:"isPublic"    form:"isPublic"`
}

// ListUpdateSchema validates ListUpdate.
var ListUpdateSchema = z.Struct[ListUpdate](z.Shape{
	"title":       z.String().Trim().Min(1).Max(80).Optional(),
	"description": z.String().Trim().Max(500).Optional(),
	"isPublic":    z.Bool().Optional(),
})

// ListItem is the add-to-list form; the note is stored on the first add only.
type ListItem struct {
	Note string `json:"note" form:"note"`
}

// ListItemSchema validates ListItem.
var ListItemSchema = z.Struct[ListItem](z.Shape{"note": z.String().Trim().Max(280).Default("")})

// ListOrder is the full new order of a list with the version it was read at.
type ListOrder struct {
	Version int      `json:"version" form:"version"`
	ItemIDs []string `json:"itemIds" form:"itemIds"`
}

// ListOrderSchema validates ListOrder; set equality is checked by library.PlanOrder.
var ListOrderSchema = z.Struct[ListOrder](z.Shape{
	"version": z.Int().Min(0),
	"itemIds": z.Slice(z.String().UUID()).Max(500),
})
