package console

import (
	"context"
	"errors"
	"flag"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// CatalogMerge is `catalog:merge --from <id> --into <id>`: the console half
// of the item merge (SDD 01 §7.5). Irreversible; the audit snapshot allows a
// manual restore.
func CatalogMerge() vukem.Command {
	var from, into string

	return vukem.CommandFunc{
		CommandName: "catalog:merge",
		Desc:        "Merge item --from into item --into (same type, irreversible)",
		DefineFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&from, "from", "", "id of the item to merge away")
			fs.StringVar(&into, "into", "", "id of the item that survives")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if from == "" || into == "" {
				return errors.New("catalog:merge needs --from and --into")
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			q, err := app.Queue()
			if err != nil {
				return err
			}

			if err := catalog.Merge(ctx, db, q, from, into, 0); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Merged %s into %s\n", from, into)

			return nil
		},
	}
}
