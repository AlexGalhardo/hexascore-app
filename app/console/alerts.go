package console

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/alerts"
)

// AlertsTest is `alerts:test [message]`: sends one message to the Telegram page chat right away and reports whether
// the Bot API accepted it (SDD 12 §7.5). It fails when TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is unset.
func AlertsTest() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "alerts:test",
		Desc:        "Send one test message to the Telegram page chat",
		Fn: func(ctx context.Context, app *vukem.App, args []string) error {
			n, err := alerts.FromApp(app)
			if err != nil {
				return err
			}

			if n == nil {
				return errors.New("alerts:test: TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are not set")
			}

			defer func() { _ = n.Close(context.WithoutCancel(ctx)) }()

			msg := strings.Join(args, " ")
			if msg == "" {
				msg = "HexaScore alerts wired (test from " + app.Config.Name + ", " + app.Now().
					Format(time.DateOnly) +
					")"
			}

			if err := n.SendPage(ctx, html.EscapeString(msg)); err != nil {
				return fmt.Errorf("alerts:test: %w", err)
			}

			_, _ = fmt.Fprintln(app.Stdout(), "Telegram accepted the message (sendMessage 200, ok:true)")

			return nil
		},
	}
}
