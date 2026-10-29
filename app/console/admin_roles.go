// Package console holds the app's custom console commands.
package console

import (
	"context"
	"errors"
	"flag"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ErrEmailArgument is returned when a command is not given exactly one email.
var ErrEmailArgument = errors.New("expected exactly one email argument")

func roleCommand(name, desc, role, done string) vukem.Command {
	return vukem.CommandFunc{
		CommandName: name,
		Desc:        desc,
		Fn: func(ctx context.Context, app *vukem.App, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("%s: %w", name, ErrEmailArgument)
			}

			acc, err := accounts.New(app)
			if err != nil {
				return err
			}

			changed, err := acc.SetRole(ctx, args[0], role)
			if errors.Is(err, orm.ErrNotFound) {
				return fmt.Errorf("%s: no user with that email", name)
			}

			if err != nil {
				return err
			}

			if !changed {
				_, _ = fmt.Fprintln(app.Stdout(), "Nothing to change:", args[0], "is already", role)

				return nil
			}

			_, _ = fmt.Fprintln(app.Stdout(), done, args[0])

			return nil
		},
	}
}

// AdminGrant is `admin:grant <email>`: gives the admin role.
func AdminGrant() vukem.Command {
	return roleCommand("admin:grant", "Grant the admin role to a user by email", models.RoleAdmin, "Granted admin to")
}

// AdminRevoke is `admin:revoke <email>`: back to the plain user role.
func AdminRevoke() vukem.Command {
	return roleCommand("admin:revoke", "Revoke the admin role from a user by email", "user", "Revoked admin from")
}

// UserReset2FA is `user:reset-2fa <email> --reason "<text>"`: clears 2FA
// after the owner verified the user's identity by hand, revokes sessions
// and tokens, records the event and mails the user.
func UserReset2FA() vukem.Command {
	var reason *string

	return vukem.CommandFunc{
		CommandName: "user:reset-2fa",
		Desc:        "Clear a user's two-factor authentication after a manual identity check",
		DefineFlags: func(fs *flag.FlagSet) { reason = fs.String("reason", "", "why the reset was approved (required)") },
		Fn: func(ctx context.Context, app *vukem.App, args []string) error {
			// vukem parses flags before the email; the documented form puts
			// --reason after it, so the remaining arguments are parsed again.
			if len(args) > 1 {
				fs := flag.NewFlagSet("user:reset-2fa", flag.ContinueOnError)
				fs.StringVar(reason, "reason", *reason, "")

				if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
					return fmt.Errorf("user:reset-2fa: %w", ErrEmailArgument)
				}

				args = args[:1]
			}

			if len(args) != 1 {
				return fmt.Errorf("user:reset-2fa: %w", ErrEmailArgument)
			}

			if *reason == "" {
				return errors.New("user:reset-2fa: --reason is required")
			}

			acc, err := accounts.New(app)
			if err != nil {
				return err
			}

			if err := acc.ResetTwoFactor(ctx, args[0], *reason); err != nil {
				if errors.Is(err, orm.ErrNotFound) {
					return errors.New("user:reset-2fa: no user with that email")
				}

				return err
			}

			_, _ = fmt.Fprintln(app.Stdout(), "Two-factor reset for", args[0])

			return nil
		},
	}
}
