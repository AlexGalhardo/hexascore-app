# Analytics and SEO access: getting the values for `.env`

Step-by-step for the owner to collect the credentials the vukem v0.36.0 integrations need (Umami, Google Analytics 4, Google Search Console, Bing Webmaster Tools, IndexNow). Framework reference: the "Analytics and Search Console" page on <https://vukem.dev/docs>.

Every variable is optional: unset means that integration is fully off (no script, no CSP change, no outbound call).

| Variable | Secret? | Where it comes from |
| --- | --- | --- |
| `UMAMI_WEBSITE_ID` | no (public in the page) | Umami, step 1 |
| `UMAMI_SCRIPT_URL` | no | Umami, step 1 (only for self-hosted) |
| `UMAMI_DOMAINS` | no | your choice: `hexascore.app,www.hexascore.app` |
| `GA_MEASUREMENT_ID` | no (public in the page) | Google Analytics, step 2 |
| `GOOGLE_SITE_VERIFICATION` | no | Search Console, step 3 (optional, the domain is already DNS-verified) |
| `BING_SITE_VERIFICATION` | no | Bing Webmaster Tools, step 4 |
| `BING_WEBMASTER_API_KEY` | **yes** | Bing Webmaster Tools, step 4 |
| `INDEXNOW_KEY` | no (published at `/{key}.txt` by design) | step 5 |

Put the values in `C:\dev\hexascore-app\hexascore-vukem\.env` for local runs. For production set them on Railway yourself (service `app`, environment `production`); for the secret one use stdin so it never lands in shell history:

```powershell
railway variable set GA_MEASUREMENT_ID=G-XXXXXXXXXX --service app --environment production --skip-deploys
"<the-key>" | railway variable set BING_WEBMASTER_API_KEY --stdin --service app --environment production --skip-deploys
```

Never paste `BING_WEBMASTER_API_KEY` in chat, in an issue or in a committed file.

## 1. Umami

1. Sign in at <https://cloud.umami.is> (or open your self-hosted Umami).
2. **Settings → Websites → Add website**. Name `HexaScore`, domain `hexascore.app`. Save.
3. Open the website → **Edit** (or the **Tracking code** tab). The snippet looks like `<script defer src="https://cloud.umami.is/script.js" data-website-id="xxxxxxxx-xxxx-...">`.
4. Copy the `data-website-id` value into `UMAMI_WEBSITE_ID`.
5. Only if self-hosted: copy the `src` URL into `UMAMI_SCRIPT_URL` (Umami Cloud is the default, leave it unset).
6. Set `UMAMI_DOMAINS=hexascore.app,www.hexascore.app` so local and preview hosts are not counted.

Check: open the site, then Umami → the website → **Realtime** shows your visit within a minute.

## 2. Google Analytics 4

1. Open <https://analytics.google.com> → **Admin** (gear icon, bottom left).
2. Create an account and a property if there is none: **Create → Property**, name `HexaScore`, time zone `America/Sao_Paulo`, currency `BRL`.
3. **Admin → Data collection and modification → Data streams → Add stream → Web**. URL `https://hexascore.app`, name `hexascore.app`. Keep **Enhanced measurement** on.
4. The stream page shows **Measurement ID** `G-XXXXXXXXXX`. Copy it into `GA_MEASUREMENT_ID`.
5. Optional, privacy: **Admin → Data collection and modification → Data retention** set to the shortest period you need; **Data collection** keep Google signals off unless you want ad features.

GA loads only after a visitor accepts the cookie banner (Consent Mode v2, LGPD). Check: accept the banner on the site, then GA → **Reports → Realtime**.

## 3. Google Search Console

The apex already has a `google-site-verification` DNS TXT record, so the domain is probably verified already.

1. Open <https://search.google.com/search-console>. If `hexascore.app` (Domain property) is listed, skip to step 4.
2. Otherwise **Add property → URL prefix** `https://hexascore.app/` → method **HTML tag**.
3. Copy only the `content` value of `<meta name="google-site-verification" content="...">` into `GOOGLE_SITE_VERIFICATION` (several tokens: comma-separated), deploy, then press **Verify**.
4. **Sitemaps** (left menu) → add `https://hexascore.app/sitemap.xml` → **Submit**.

Check: the sitemap shows status **Success** after Google fetches it (can take hours).

## 4. Bing Webmaster Tools

1. Open <https://www.bing.com/webmasters> and sign in.
2. **Add site**. Easiest: **Import from Google Search Console** (verifies with no token). Or add `https://hexascore.app/` manually → method **HTML Meta Tag** → copy the `content` of `<meta name="msvalidate.01" content="...">` into `BING_SITE_VERIFICATION`, deploy, then **Verify**.
3. **Sitemaps** → submit `https://hexascore.app/sitemap.xml`.
4. API key: **Settings** (gear, top right) → **API access** → **API Key** → **Generate**. Copy it into `BING_WEBMASTER_API_KEY`. This one is a secret.

Check: after deploy, the vukem Bing client calls return data for the site (query and page stats need a few days of traffic).

## 5. IndexNow

No account needed. Bing, Yandex, Seznam, Naver and others share submissions; Google does not use IndexNow (it reads the sitemap).

1. Generate a key: 32 hex characters, for example in PowerShell:

   ```powershell
   -join ((1..32) | ForEach-Object { '{0:x}' -f (Get-Random -Maximum 16) })
   ```

   Or ask the orchestrator to generate one. Allowed: 8 to 128 characters of `a-z`, `A-Z`, `0-9`, `-`.
2. Put it in `INDEXNOW_KEY`. The app serves it at `https://hexascore.app/{key}.txt`; that public file is how search engines confirm the submissions are yours.
3. Nothing to register: the first submission after deploy activates it.

Check: `https://hexascore.app/{key}.txt` returns the key; Bing Webmaster Tools → **IndexNow** lists the submitted URLs after an item is created or edited.

## After filling the values

Tell the orchestrator which ones are set (names only, never the secret values). The ADOPT-vukem-0.36.0 row wires them in, adds the LGPD cookie banner for GA, and confirms each check above.
