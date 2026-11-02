# Cloudflare: steps for hexascore.app

Do the steps in the order below. Part 1 has no risk. Parts 2 and 3 change the production DNS, so read everything before you start.

---

## Part 1: Turnstile (now, no risk)

1. In the left menu, open Turnstile and click **Add widget**.
2. Fill in:
   - Widget name: `HexaScore`
   - Hostnames: `hexascore.app` and `www.hexascore.app`. Do not include `localhost`, because development uses Cloudflare's official test keys.
   - Widget mode: **Managed** (recommended). It only shows a challenge when there is suspicion.
   - Pre-clearance: **No**. It only makes sense with the site behind the Cloudflare proxy.
3. Click **Create**. Two keys appear:
   - **Site Key**: public. You can send it here in the chat.
   - **Secret Key**: secret. Do not send it in the chat. Put it straight into the Railway dashboard: service `app` > Variables > New Variable, name `TURNSTILE_SECRET_KEY`, pasting the value there.

---

## Part 2: Move the domain to Cloudflare (D-28, ROADMAP item P2.A.2.2)

Today the DNS is on Vercel (`ns1/ns2.vercel-dns.com`). D-28 planned the move to Cloudflare with the proxy turned off (grey cloud). This part changes the production DNS: one forgotten record takes down the site or the e-mail.

1. Open **Account Home** > **Add a domain** > `hexascore.app` > **Quick scan for DNS records** > Free plan.
2. On the review screen, check that each of these records exists, all with the grey cloud ("DNS only"):

| Type  | Name                | Value                                                          |
| ----- | ------------------- | -------------------------------------------------------------- |
| CNAME | @ (apex)            | the `*.up.railway.app` target of the apex (see the note below) |
| CNAME | www                 | `m9rebj3k.up.railway.app`                                      |
| CNAME | api                 | `8niu0qtr.up.railway.app`                                      |
| TXT   | _railway-verify     | copy from Vercel                                               |
| TXT   | _railway-verify.www | copy from Vercel                                               |
| TXT   | _railway-verify.api | copy from Vercel                                               |
| TXT   | @                   | `google-site-verification=...` (copy from Vercel)              |
| MX    | send                | `feedback-smtp.us-east-1.amazonses.com`, priority 10           |
| TXT   | send                | `v=spf1 include:amazonses.com ~all`                            |
| TXT   | resend._domainkey   | the long DKIM key (copy from Vercel)                           |

   **Careful with the apex:** the scan usually imports the apex as an `A 69.46.46.87` record, which is Vercel's ALIAS resolved. Delete that `A` record and create a `CNAME @` pointing to the Railway target. The value is in Railway (service `app` > Settings > Networking > `hexascore.app`) or in the ALIAS record of `@` in the Vercel DNS. Cloudflare does the "CNAME flattening" on its own.
3. In **SSL/TLS** > **Overview**, choose **Full (strict)**. With the proxy off this changes nothing yet, but it is correct for later.
4. Cloudflare shows 2 nameservers (`xxx.ns.cloudflare.com`). Stop here and tell me before you change anything. I check your record list against the one that exists today.
5. After my OK, change the nameservers where the domain is registered. If it was bought on Vercel, it is in **Vercel** > **Domains** > `hexascore.app` > **Nameservers**.
6. Propagation takes from minutes to a few hours. I follow it with `dig` and check the site, the API, `/up` and the e-mail.

---

## Part 3: E-mail for support@ and privacy@ (only after Part 2 is active)

This resolves items 19 to 21 of `questions-to-dev.md`: today nobody can receive e-mail sent to these addresses.

1. **Email** > **Email Routing** > **Get started**. Cloudflare creates the apex MX and SPF records by itself.
2. **Destination addresses**: add your personal e-mail and confirm the link that arrives in it.
3. **Routing rules**: `support@hexascore.app` to your e-mail, and `privacy@hexascore.app` to your e-mail.
4. **DNS** > **Add record** > TXT, name `_dmarc`, value `v=DMARC1; p=none; rua=mailto:support@hexascore.app`.

With this you receive the e-mails. To reply as `support@`, Email Routing alone is not enough: you can configure Gmail to "send as" through the Resend SMTP. I can explain it later, if you want.

---

## What I need you to send me

1. The Turnstile **Site Key** (public).
2. Confirmation that `TURNSTILE_SECRET_KEY` is in Railway, without the value.
3. For Part 2, the warning before the nameserver change, with a screenshot of the Cloudflare DNS records screen.

If you only want Turnstile today, do Part 1 and stop. Parts 2 and 3 can wait until you have time to follow the DNS change with me.

---

Yes, Cloudflare has all three things, and the path I recommend is the official MCP with OAuth. I confirmed this in the official docs.

| Option                                    | What it is                                                                                                                                                                                            | Fits here?        |
| ----------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| Cloudflare API MCP (official)             | Remote server at `https://mcp.cloudflare.com/mcp`. Covers the whole API (more than 2,500 endpoints: DNS, zones, Turnstile, Email Routing and so on) with only 3 tools: `search`, `execute` and `docs` | Yes, recommended  |
| REST API (`api.cloudflare.com/client/v4`) | For example, Turnstile widgets are at `/accounts/{id}/challenges/widgets`, with the permission `Turnstile Sites Write`                                                                                | Yes, with a token |
| Wrangler (CLI)                            | Focused on Workers, R2 and D1. Does not handle DNS or Turnstile well                                                                                                                                  | Not for this      |

Authentication:

- **OAuth (recommended):** you connect, Cloudflare opens in the browser, and you choose the permissions. No token goes through the chat. If a permission is missing, the MCP itself asks you to authorise again.
- **API token:** only for CI automation. We do not need it now.

So that I can automate, run this in your terminal (outside here) and then restart Claude Code:

```bash
claude mcp add --transport http cloudflare https://mcp.cloudflare.com/mcp
```

Then type `/mcp` here, choose `cloudflare` and authorise in the browser.

What I will do on my own:

- create the Turnstile widget and configure hostnames and mode;
- create the `hexascore.app` zone and all the DNS records in the Part 2 table, with the apex `CNAME` fixed and the proxy off;
- adjust SSL/TLS and create the DMARC record;
- turn on Email Routing with the rules for `support@` and `privacy@`.

What stays with you:

- Changing the nameservers on Vercel, where the domain is registered. Before that, I show you the comparison of the records.
- Clicking the verification link that Cloudflare sends to your personal e-mail, in Email Routing.
- Putting `TURNSTILE_SECRET_KEY` into Railway. The permission classifier blocks me from writing secrets. I create the widget without bringing the secret into the chat, and you copy it from the Cloudflare dashboard straight into the Railway dashboard.

Meanwhile, the specialist reviews the anti-scraping work, and the vukem workers close the analytics and Turnstile PRs.

Sources:

- Cloudflare own MCP servers: <https://developers.cloudflare.com/agents/model-context-protocol/mcp-servers-for-cloudflare/>
- cloudflare/mcp on GitHub: <https://github.com/cloudflare/mcp>
- Turnstile: managing widgets through the API: <https://developers.cloudflare.com/turnstile/get-started/widget-management/api/>
