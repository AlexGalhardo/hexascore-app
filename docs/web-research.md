# Competitive research: IMDb, TMDB, IGDB, Metacritic, Rotten Tomatoes, Goodreads, Skoob

Phase 1 deliverable. Research date: 2026-10-28. Method: public documentation, official API pages, terms of use, engineering blogs, press and papers only; no scraping, no authenticated pages. Each competitor section was produced by a research agent and reviewed by the architect; every claim carries a numbered source, and anything not confirmed by a source is tagged `[NOT VERIFIED]`.

How to read this document:

- Sections 1–7: one per competitor, same six headings each (presentation, API, rating calculation, user interaction, data freshness, weaknesses/opportunities).
- Section 8: comparative matrix.
- Section 9: implications for HexaScore (the only section with opinions).
- Section 10: consolidated sources.

## Executive summary

1. **No competitor covers all four media.** IMDb, TMDB, Metacritic and Rotten Tomatoes stop at film/TV (IMDb's paid bulk data includes games but the product does not); IGDB is games only; Goodreads and Skoob are books only. A single identity model across media is open ground.
2. **Every incumbent hides its score formula.** IMDb ("weighted average", undisclosed), Metacritic (per-publication weights, undisclosed), TMDB (no documented formula), IGDB (no documented formula), Goodreads and Skoob (plain means, undocumented). Only Rotten Tomatoes publishes exact rules, and its headline score is binary. A published, versioned formula with confidence and minimum counts is a differentiator nobody occupies.
3. **Review bombing is handled reactively everywhere.** IMDb applies an undisclosed alternate weighting; Metacritic added a 36-hour wait (2020) and manual moderation (2023); Goodreads throttles ratings during unusual activity (2023); Rotten Tomatoes removed pre-release scores and ties verification to Fandango tickets (2019). None publishes the detection rules.
4. **Developer access is closed or expensive.** IMDb: AWS Data Exchange, US$150,000/year listed. Rotten Tomatoes and Metacritic: licensing via proposal forms or Fabric Origin, no public prices. Goodreads: API retired in 2020. Skoob: no API. TMDB: free for non-commercial, commercial by negotiated written agreement, 6-month cache limit. IGDB: free including commercial use with a partnership and visible attribution, 4 rps, partner-only dumps. A self-serve, priced, documented API with a clear per-field licence is a differentiator; it is also HexaScore's main legal exposure (see 9.1).
5. **Community editing exists only at TMDB and Goodreads (librarians) and IMDb (staff-reviewed, 1–7 days).** Metacritic, Rotten Tomatoes, IGDB (staff-validated) and Skoob give users little or no catalogue control. A moderated suggestion flow with public status and audit trail beats all of them if moderation latency stays under a day.
6. **Freshness tooling worth copying**: TMDB daily ID exports and `changes` endpoints; IGDB webhooks, `updated_at`/`checksum` and PopScore primitives; IMDb daily TSV refresh.
7. **Brazil**: Skoob has 10–11 million registered users and 1.3 million monthly active users (May 2025), is books-only, has no API, paused groups and an owner (Skeelo) focused on ebooks. Publisher relations (Cortesias) are its moat; UI and openness are not.

## 1. IMDb

Citation numbers in this section refer to the "Sources (IMDb)" list at its end.

### 1.1 Information presentation

- **Title page**: the displayed rating is a weighted average of IMDb users and is hidden until a title has at least five ratings [2]. Critic material is separate: a Metascore where available plus links to professional reviews [2]. Full hierarchy (poster, cast, trivia order) [NOT VERIFIED].
- **Ratings breakdown** is always shown as a distribution of votes [2]. Demographic (age/gender) breakdowns were removed; a top-five country filter and a per-episode heat map replaced them [2].
- **Advanced Search**: reachable via "All" then "Advanced Search" [11]. Documented filters: title type, genre, year range, user rating, own ratings; sort by Popularity, A-Z, Year, User rating, Your rating [11]. Exclusion by prefixing a parameter value with `!` in the URL [11]. Full filter list [NOT VERIFIED].
- **Charts**: Top 250 Movies includes only theatrical feature films and only "regular" user ratings [2]. Top 250 TV added in 2015; Bottom 100 requires 10,000 votes [10]. "Popular" = MOVIEmeter, based on title page views, updated weekly, a public-interest signal rather than quality [16]. The bulk file covers the top 2,000 movies [17].
- **Person pages**: STARmeter is the people equivalent of MOVIEmeter [16]. Layout [NOT VERIFIED].
- **Reviews vs critics**: the overall score uses only IMDb users; IMDb states it does not consider published reviews or critics' ratings [2]. User reviews are shown in a "featured" order meant to represent varied opinions, not only positive ones [18].
- **Mobile app** [NOT VERIFIED].

### 1.2 API and data access

- **Non-Commercial Datasets**: seven gzipped UTF-8 TSV files (`title.akas`, `title.basics`, `title.crew`, `title.episode`, `title.principals`, `title.ratings`, `name.basics`) at datasets.imdbws.com [1]. Licence: "available for access to customers for personal and non-commercial use"; "refreshed daily" [1]. No API, rate limit or price on that page [1]. The developer domain redirects to data.imdb.com [1].
- **Official API**: "exclusively available through AWS Data Exchange." Requires an AWS account and access keys; subscription requests are "processed by IMDb staff within five business days"; the API key arrives by email [4]. Pricing and rate limits are not stated on the getting-access page [4].
- **Essential Metadata (AWS Marketplace)**: covers movies, TV/OTT, video-game titles, performers and creators, including the 1–10 rating. Single dimension "Product Access (Units)"; listed "12-month contract" at "$150,000.00"; "Pricing is based on the duration and terms of your contract"; "Refunds are not offered"; delivered as JSON files on S3 [5]. A third-party vendor repeats the ~$150k/year figure [19, unverified against IMDb].
- **IMDbPro** pricing [NOT VERIFIED]: pro.imdb.com returned 403 to the research agent.
- **Scraping**: the Conditions of Use prohibit data mining, robots and screen scraping without express written consent [3]. Wording taken from search-result mirrors; the live page was not fetched [3, NOT VERIFIED on live page].

### 1.3 Rating calculation

- Displayed score is "weighted rating averages rather than raw data averages." Exact weighting undisclosed [2].
- Top 250 formula published by IMDb: `WR = (v ÷ (v+m)) × R + (m ÷ (v+m)) × C`, with `m = 25,000` currently and `C` the mean across the chart [2]. Only "regular" IMDb users count; the criteria are not disclosed [2, 10].
- Earlier thresholds (3,000 votes; 1,300 in blogs) are superseded or unreliable [2, not verified as history].
- Ratings count once the account is verified; one rating per user per title, overwritten on re-rate [2].
- Vote stuffing: IMDb says it "has several safeguards in place to automatically detect and defeat" ballot stuffing, keeps valid ratings, and applies an "alternate weighting calculation" [2]. Case: *The Little Mermaid* (2023) showed a notice of "unusual voting activity"; about 13,000 of 33,000+ reviews (~39%) were 1-star, yet the displayed score stayed 7/10 [14, 15]. The method was not disclosed [14].

### 1.4 User interaction

- **Ratings**: 1–10, editable, deletable [2].
- **Reviews**: minimum 300 characters, English only, no markup, must be about the title; headline, rating, optional spoiler flag; reports remove only guideline violations [18].
- **Check-ins**: public, with a short comment, not available for unreleased titles [12]. Watchlist requires registration [13].
- **Contributions**: "Edit Page" button under "Contribute to this page" [9]. Title images, reviews and parental guides use their own flows; new titles use the New Title Form [9]. Status tracked at contribute.imdb.com for new titles, episodes, trivia and images [9].
- **Contributor levels**: none documented. The charter says "IMDb tracks each contributor's accuracy over time," and inaccurate submitters face "increasing levels of additional proof" [8].
- **Badges**: contributing can earn badges [13]; criteria [NOT VERIFIED].
- **Lists**: user lists exist; details [NOT VERIFIED beyond lists/check-ins].

### 1.5 Data freshness and sourcing

- Founded 17 Oct 1990; Amazon subsidiary since 1998 [10]. About 25.9 million titles and 14.8 million person records as of Sept 2025 [10].
- Community and studio contributions go to a data editing team; IMDb "reserves the right to reject or delete information" [9, search summary; live page not fully read].
- Approved changes appear "within 1 - 7 days" depending on data type; not processed in order received; a Processing Times page is updated every business day [7, 8].
- Dataset refresh: daily [1]. IMDbPro/studio sourcing details [NOT VERIFIED].

### 1.6 Weaknesses and opportunities

- **Weaknesses (documented)**: review bombing and undisclosed weighting [2, 14]; contribution latency of 1–7 days [7]; API only via AWS at roughly $150k/year minimum listed [5]; commercial use requires a separate licence [1, 3]; demographic breakdowns removed [2]; no self-serve public API [4]; games are in the bulk product but there is no native game/book experience [5]; books [NOT VERIFIED].
- **Opportunities**: publish the scoring method and anti-manipulation signals; show per-cohort distributions as an opt-in; cheap or tiered developer access with clear licensing; faster moderation with status tracking per edit; one catalog across movies, series, games and books, which IMDb's product does not do.

### Sources (IMDb)

1. <https://data.imdb.com/non-commercial-datasets/> (via developer.imdb.com redirect), accessed 2026-10-28
2. <https://help.imdb.com/article/imdb/track-movies-tv/ratings-faq/G67Y87TFYYP6TWAV>, accessed 2026-10-28
3. <https://m.imdb.com/conditions> (search mirror; live page not fetched), accessed 2026-10-28
4. <https://data.imdb.com/documentation/api-documentation/getting-access/>, accessed 2026-10-28
5. <https://aws.amazon.com/marketplace/pp/IMDb-MovieTVOTT-Data/prodview-yeuyizioqmfsy>, accessed 2026-10-28
6. <https://help.imdb.com/article/contribution/contribution-information/when-will-my-update-be-added/GEKS9SBHDEHBAVC7>, accessed 2026-10-28
7. <https://contribute.imdb.com/charter>, accessed 2026-10-28
8. <https://help.imdb.com/article/contribution/contribution-information/adding-data/G6BXD2JFDCCETUF4> (search summary), accessed 2026-10-28
9. <https://en.wikipedia.org/wiki/IMDb>, accessed 2026-10-28
10. <https://help.imdb.com/article/imdb/discover-watch/using-the-advanced-search-feature/GLUEUYWPQNPTEVPU> (search summary), accessed 2026-10-28
11. <https://help.imdb.com/article/imdb/track-movies-tv/check-ins-faq/GG59ELYW45FMC7J3> (search summary), accessed 2026-10-28
12. <https://help.imdb.com/article/imdb/general-information/why-should-i-register-on-imdb/GHB62T7USTMYMCDC> (search summary), accessed 2026-10-28
13. <https://fortune.com/2023/05/30/imdb-revamps-rating-system-the-little-mermaid>, accessed 2026-10-28
14. <https://thegrio.com/2023/06/01/imdb-updates-rating-system-the-little-mermaid-negative-review-bombing/> (search summary), accessed 2026-10-28
15. <https://help.imdb.com/article/imdbpro/industry-research/faq-for-starmeter-moviemeter-and-companymeter/GSPB7HDNPKVT5VHC> (search summary), accessed 2026-10-28
16. <https://developer.imdb.com/documentation/bulk-data-documentation/data-dictionary/meters/> (search summary), accessed 2026-10-28
17. <https://help.imdb.com/article/contribution/contribution-information/user-review-guidelines/GABTWSNLDNFLPRRH> (search summary), accessed 2026-10-28
18. <https://api.market/blog/sleeyax/entertainment/best-movie-apis> (third-party, unverified), accessed 2026-10-28

(Number 6 intentionally unused; source 2 also covers the demographic-breakdown removal.)

## 2. TMDB (The Movie Database)

Citation numbers refer to "Sources (TMDB)". Caveat from the research agent: quotes came through a summarising fetch tool; verify wording verbatim before any legal decision. Several official pages returned 404 (contribution bible root, pagination, discover caveats), so those items are `[NOT VERIFIED]`.

### 2.1 Information presentation

- Page layout (backdrop, poster, user-score circle, overview, cast, media, recommendations, keywords), discover UI, person pages, lists and mobile apps: no official page describing them was retrieved `[NOT VERIFIED]`.
- Discover is exposed in the API as "Find movies using over 30 filters and sort options" [4].

### 2.2 API and data access

- **Auth**: v3 accepts "a single query parameter, `api_key`, or by using your access token as a `Bearer` token". The v4 Read Access Token is "the default method to authenticate", sent in the `Authorization` header. "Both authentication methods provide the same level of access" [5]. v4 user tokens (lists, watchlist, ratings) `[NOT VERIFIED]`.
- **Endpoint families seen**: `GET /3/discover/movie` (default `sort_by=popularity.desc`, 14 sort values, `include_adult` default false, many `with_*` and `watch_region` filters) [4], `GET /3/movie/{movie_id}/changes` [9], `/configuration` and image paths [7]. Search and find-by-external-id were not fetched `[NOT VERIFIED]`.
- **Pagination**: `page` default 1. The docs give no maximum page or page size; the common "20 per page / max 500" figure is `[NOT VERIFIED]`.
- **Rate limit**: "sit somewhere in the 40 requests per second range" and "could change at any time". The old limit of 40 requests per 10 seconds was disabled on 2019-12-16. Per-IP scope and headers are not documented. Clients should "respect the `429`" [2].
- **Image CDN**: URL = `base_url` + `file_size` + `file_path`, the first two from `/configuration`; example host `https://image.tmdb.org/t/p/`; sizes such as `w500` and `original` appear in examples; SVG logos should use `original` [7].
- **Daily ID exports**: `https://files.tmdb.org/p/exports/movie_ids_MM_DD_YYYY.json.gz` and equivalents for TV, people, collections, networks, keywords and companies; published daily from about 07:00 UTC, all by 08:00 UTC; adult sets separate; kept "only for 3 months"; "there is currently no authentication" but that "could change"; they are "not, nor intended to be full data exports" [3].
- **Changes API**: `/movie/{id}/changes` returns the last 24 hours by default; up to 14 days per query with `start_date`/`end_date` [9]. The bulk `/movie/changes` list endpoint was not fetched `[NOT VERIFIED]`.
- **Versioning**: v3 and v4 coexist; change notices go to the docs and a Trello board [8]. SDKs `[NOT VERIFIED]`.
- **Developer experience**: OpenAPI-based reference with per-endpoint pages. The FAQ states no SLA, only "every reasonable attempt" [8].
- **Pricing and commercial terms**:
  - Non-commercial use is free "as long as you attribute TMDB as the source of the data and/or images". A project is commercial if "the primary purpose is to create revenue" [8].
  - Commercial access: email <sales@themoviedb.org> with subject "Commercial API Use" plus country; the API-for-business page shows no prices [10].
  - API Terms of Use (last updated 2023-10-20) [1]: the licence "does not permit any commercial use of TMDB, the TMDB APIs, or TMDB Content"; commercial use "is only permitted under a separate written agreement".
  - Attribution: "You must use the TMDB logo"; a not-endorsed notice is required [1].
  - Caching: you may not "Cache, for longer than 6 months, any information obtained through or from TMDB"; delete on termination [1].
  - Redistribution: licence is "non-exclusive, non-transferable, non-sublicensable"; resale of "TMDB Content" is barred [1].
  - AI and derivatives: the terms bar training ML/AI on TMDB data, creating derivatives, and providing "archived or cached data sets containing TMDB Content" for training [1].
  - No numeric rate limit in the terms; bandwidth abuse is banned [1].

### 2.3 Rating calculation

- **Displayed score**: the official percentage formula was not found `[NOT VERIFIED]`. Third-party sites assume it is the 10-point average × 10 (unverified).
- **`vote_average` / `vote_count`**: no official source says the average is weighted; weighted-rating formulas appear only in third-party blogs. Minimum-vote thresholds are not documented `[NOT VERIFIED]`.
- **Popularity**: "a 'lifetime' popularity score". Inputs differ per type. Movies: "Number of votes for the day" and "Release date"; people: "Number of views for the day" and "Previous days score". "Each model builds their popularity value slightly differently"; no formula; recalculation frequency not stated [6].
- **Trending**: movies, TV and people; "time windows are much shorter (daily, weekly)"; no formula [6].
- **Anti-manipulation**: not documented `[NOT VERIFIED]`.

### 2.4 User interaction

- **Contribution rules** ("New content" bible page) [11]: moderators "are tasked to remove unsupported content". Video games are not supported. Amateur films need one of five criteria. Removed creators get a private message saying the entry "did not meet our new content guidelines" and can contest with "at least one independent online source". "New people entries with no credits or incorrect credits will be deleted."
- Edit locking, change log, leaderboards, reviews, discussions, watchlist, favourites: pages returned 404 or were not retrieved `[NOT VERIFIED]`.
- **Contributor licence** (Terms of Use, last updated 2024-05-10) [12]: "We do not claim any ownership right in Your Contributions"; contributions are licensed "irrevocable, perpetual, non-exclusive" with a machine-learning training clause; users may not "scrape" or train AI on other users' contributions "unless expressly authorized by Us in writing".

### 2.5 Data freshness and sourcing

- Community-driven with moderators enforcing creation rules [11].
- Freshness signals: daily ID exports [3], 24 h-default changes endpoint with 14-day window [9].
- Image, locale and translation coverage `[NOT VERIFIED]`.

### 2.6 Weaknesses and opportunities

- **Criticism**: removals described as opaque on community forums (`[NOT VERIFIED]` as official). No games or books; games explicitly excluded [11].
- **Rating reliability**: small `vote_count` values are unhandled in the docs; a competitor can present a Bayesian or minimum-vote score with confidence.
- **Commercial opacity**: no public price list; a negotiated contract.
- **Opportunities**: documented, fast-moving catalogue for games and books; own rating scheme with confidence intervals; licence-compliant ingestion; own change feed for API consumers.
- **Legal constraints for HexaScore (not legal advice)**: the API terms bar commercial use without a written agreement; caching beyond 6 months and resale are barred; AI training and redistribution of cached sets are barred; logo attribution and not-endorsed notice are mandatory. The paid HexaScore API needs a written commercial agreement with TMDB, or must stop serving TMDB-derived fields, before launch.

### Sources (TMDB)

1. <https://www.themoviedb.org/api-terms-of-use> (last updated 2023-10-20), accessed 2026-10-28
2. <https://developer.themoviedb.org/docs/rate-limiting>, accessed 2026-10-28
3. <https://developer.themoviedb.org/docs/daily-id-exports>, accessed 2026-10-28
4. <https://developer.themoviedb.org/reference/discover-movie>, accessed 2026-10-28
5. <https://developer.themoviedb.org/docs/authentication-application>, accessed 2026-10-28
6. <https://developer.themoviedb.org/docs/popularity-and-trending>, accessed 2026-10-28
7. <https://developer.themoviedb.org/docs/image-basics>, accessed 2026-10-28
8. <https://developer.themoviedb.org/docs/faq> (page metadata 2025-10-07), accessed 2026-10-28
9. <https://developer.themoviedb.org/reference/movie-changes>, accessed 2026-10-28
10. <https://www.themoviedb.org/api-for-business>, accessed 2026-10-28
11. <https://www.themoviedb.org/bible/new_content>, accessed 2026-10-28
12. <https://www.themoviedb.org/terms-of-use> (last updated 2024-05-10), accessed 2026-10-28

## 3. IGDB

Citation numbers refer to "Sources (IGDB)". Research status: the research agent's fetcher got HTTP 403 from igdb.com and dev.twitch.tv; the architect then retrieved the official API documentation (api-docs.igdb.com) with a different fetcher on 2026-10-28, so sections 3.2, 3.3 and 3.5 are verified against the official text. igdb.com/api, igdb.com/terms_of_service and the Twitch Developer Services Agreement body still could not be read (403 or empty body); items that depend on them stay `[NOT VERIFIED]`.

### 3.1 Information presentation

- **Game page** (official docs describe the data, not the page): the `games` entity exposes cover, artworks, screenshots, videos, summary, storyline, platforms, release dates, genres, themes, game modes, player perspectives, keywords, franchises, collections, involved companies, game engines, language supports, DLCs, expansions, remakes, remasters, ports, bundles, similar games, external games (store IDs), websites, age ratings, and the rating fields in 3.3 [3]. Page layout itself `[NOT VERIFIED]`.
- Region-specific cover art supported as of May 2025 [4]. Backloggd said it was waiting on IGDB's new genre system [4].
- **Community scale** (Wikipedia, March 2023, "citation needed"): 99,000 members, 428,000 games including 196,500 DLCs and re-releases, 44,300 companies, 217,000 people [1]. Stale.
- Discovery pages, release calendar, filters, company/franchise pages, mobile app: `[NOT VERIFIED]`.

### 3.2 API and data access

- **Account**: a Twitch account with two-factor authentication, an application registered in the Twitch Developer Portal with Client Type `Confidential` (the OAuth redirect URL is unused; `localhost` is suggested) [3].
- **Auth**: `POST https://id.twitch.tv/oauth2/token?client_id=…&client_secret=…&grant_type=client_credentials`; the response carries `access_token`, `token_type: bearer` and `expires_in` in seconds (example value 5,587,808 s, about 64 days); the token "must be refreshed" when it expires. Requests send `Client-ID` and `Authorization: Bearer` headers [3].
- **Request format**: `POST https://api.igdb.com/v4/{endpoint}` with an APICalypse body (`fields`, `where`, `sort`, `search`, `limit`/`l`, `offset`/`o`, expanders such as `fields screenshots.*`) [3].
- **Pagination**: default `limit` 10, maximum 500; `offset` for paging [3].
- **Rate limits**: "There is a rate limit of **4 requests per second**"; over it, `429 Too Many Requests`; "You are able to have up to 8 *open* requests at any moment in time" [3].
- **Images**: `https://images.igdb.com/igdb/image/upload/t_{size}/{hash}.jpg`; the API returns `t_thumb` URLs by default and the client builds other sizes from `image_id`; documented sizes include `cover_small` 90×128, `cover_big` 264×374, `logo_med` 284×160, `screenshot_med` 569×320, `screenshot_big` 889×500, `screenshot_huge` 1280×720; `_2x` suffix for retina [3].
- **Protocol Buffers**: append `.pb` to any request (`/v4/games.pb`); the proto3 schema of the whole API is published at `https://api.igdb.com/v4/igdbapi.proto` [3].
- **Webhooks**: `POST /v4/{endpoint}/webhooks` with form fields `url`, `method` (`create`, `update`, `delete`) and `secret`; every delivery carries `X-Secret`, `X-Endpoint`, `X-Operation` and `User-Agent: IGDB-Webhook-Bot`; deliveries do not come from a static IP, so verify by secret, not by allowlist; `DELETE /v4/webhooks/{id}` removes one [3].
- **Data dumps**: every endpoint is available as a daily CSV dump "exclusively available to our Data Partners"; `GET /v4/dumps/{endpoint}` returns a presigned S3 URL valid for 5 minutes plus `schema_version` and the column schema [3].
- **Versioning**: all documented paths are `/v4/`; deprecated fields are kept with a `DEPRECATED!` note (e.g. `category` → `game_type`, `collection` → `collections`, `follows` to be removed) [3]. Changelog page `[NOT VERIFIED]`.
- **Wrappers**: the docs list community wrappers plus a third-party OpenAPI document and Postman collection [3].
- **Licence and price** (official FAQ, quoted) [3]:
  - "The IGDB.com API is free for **non-commercial** usage under the terms of the Twitch Developer Service Agreement."
  - Business FAQ 1: "Yes, we offer commercial partnerships for users looking to integrate the API in monetized products. From our side, as part of the partnership, we ask for user facing attribution to IGDB.com from products integrating the IGDB API." Contact: <partner@igdb.com>.
  - Business FAQ 2: "The API is free for both non-commercial and commercial projects."
  - Business FAQ 3 (storing data): "Yes. In fact, we prefer if you store and serve the data to your end users."
  - Business FAQ 4 (attribution): "We expect fair attribution, i.e. attribution that is visible to your users and located in a static location (e.g. not in a change log)."
  - Code samples are "Program Materials" under the Twitch Developer Services Agreement [3].
  - The Twitch Developer Services Agreement text itself (data retention, redistribution, sublicensing limits) could not be read [10] `[NOT VERIFIED]`. A forum post quotes a 24-hour cache limit [9] `[NOT VERIFIED]`, which conflicts with Business FAQ 3; the agreement wins over the FAQ in a dispute, so it must be read.

### 3.3 Rating calculation

- Documented fields on `games` [3]: `rating` "Average IGDB user rating", `rating_count`; `aggregated_rating` "Rating based on external critic scores", `aggregated_rating_count`; `total_rating` and `total_rating_count` (combination of the two; exact formula not stated); `hypes` (pre-release follows). No weighting, minimum-vote rule or anti-manipulation measure is documented `[NOT VERIFIED]`.
- **PopScore / popularity primitives** [3]: `popularity_types` lists sources and types; documented primitives: IGDB Visits, IGDB Want to Play, IGDB Playing, IGDB Played, Steam 24hr Peak Players, Steam Positive/Negative/Total Reviews, Steam Global Top Sellers, Most Wishlisted Upcoming, Twitch Hours Watched (past 24 h). `popularity_primitives` rows carry `game_id`, `popularity_type`, `value`, `calculated_at`. The docs explicitly encourage consumers to "Define your own trend metrics" by weighting primitives (example: 0.6 × Want to Play + 0.4 × Playing).
- **Time to beat**: `game_time_to_beats` exposes `hastily`, `normally`, `completely` (seconds, averages) and `count` of submissions [3].
- Comparators for context: OpenCritic uses an arithmetic mean of normalised top-critic scores [11]; Metacritic uses undisclosed per-publication weights [12].

### 3.4 User interaction

- Registration optional and free; registered users rate, list and review games; users can edit or create pages, but IGDB staff validate changes before publication [1].
- The PopScore primitives confirm "Want to Play", "Playing" and "Played" lists exist on igdb.com [3].
- Contribution workflow, moderator roles, karma, verified contributors, release-date crowdsourcing: `[NOT VERIFIED]`.

### 3.5 Data freshness and sourcing

- Community submissions with staff validation [1].
- Every entity has `created_at`, `updated_at` and a `checksum` field for change detection; webhooks push `create`/`update`/`delete` events; partner dumps refresh daily "(within 24 hours)" [3].
- Store and external IDs via the `external_games` endpoint [3]; the per-store category list `[NOT VERIFIED]`.
- After Twitch moved game data to IGDB, some legacy Twitch endpoints returned 0 for `giantbomb_id` [8].

### 3.6 Weaknesses and opportunities

- **Weaknesses**: auth and account tied to Twitch developer accounts (2FA, confidential app); 4 rps / 8 in-flight is low for bulk sync without partner dumps; commercial partnership is informal (email) and the governing legal text is hard to read; rating formula undocumented; community size numbers stale; no movies or books.
- **Opportunities**: own rating signal (weighted, anti-manipulation documented) instead of mirroring IGDB's; cover movies and books; publish own rate-limit and licence statement; use IGDB webhooks plus `updated_at` for incremental sync instead of full re-pulls; build HexaScore's "trending" from own primitives, as IGDB's docs recommend for theirs.
- **Legal constraints if reusing IGDB commercially (not legal advice)**: the official FAQ allows commercial use and local storage and asks for visible static attribution plus a partnership arranged through <partner@igdb.com>; the Twitch Developer Services Agreement is the binding document and was not read. Action before launch: email <partner@igdb.com> describing HexaScore (including the paid API that re-serves IGDB-derived fields) and keep the written answer; read the agreement verbatim. Until then IGDB-derived fields are display-only on the site and excluded from the paid API, as `legacy/PLAN.md` §2 intended.

### Sources (IGDB)

1. <https://en.wikipedia.org/wiki/IGDB> (fetched), accessed 2026-10-28
2. <https://discuss.dev.twitch.com/t/commercial-use-of-igdb-api/23567> (fetched; 2020-01-06 staff reply), accessed 2026-10-28
3. <https://api-docs.igdb.com/> (official API documentation: Getting Started, Rate Limits, Pagination, Images, Protocol Buffers, Webhooks, Partnership/Data Dumps, PopScore, Game and Game Time To Beat endpoints, FAQ, License), fetched and indexed by the architect, accessed 2026-10-28
4. <https://backloggd.medium.com/developer-update-may-2025-db5c66ffa6cc> (search snippet), accessed 2026-10-28
5. <https://discuss.dev.twitch.com/t/response-headers-oauth-access-tokens-getting-rate-limited/46963> (fetched; 2023-08-17 replies), accessed 2026-10-28
6. <https://apis.io/rules/igdb/igdb-rules/> (unofficial ruleset, last modified 2026-04-28), accessed 2026-10-28
7. <https://www.igdb.com/api> (HTTP 403 on both fetchers), accessed 2026-10-28
8. <https://discuss.dev.twitch.com/t/legacy-endpoints-returning-0-for-giantbomb-id-following-igdb-migration/29125> (search snippet), accessed 2026-10-28
9. <https://blog.twitch.tv/en/2016/08/24/new-developer-agreement-and-terms-of-service-5227c1ebf0fa/> (search snippet), accessed 2026-10-28
10. <https://legal.twitch.com/en/legal/developer-agreement/> (fetched twice, no body text), accessed 2026-10-28
11. <https://en.wikipedia.org/wiki/OpenCritic> (search snippet), accessed 2026-10-28
12. <https://www.gamedeveloper.com/business/full-sail-study-attempts-to-shed-light-on-metacritic-s-weighting-system> (search snippet), accessed 2026-10-28
13. <https://www.igdb.com/terms_of_service> (HTTP 403), accessed 2026-10-28

## 4. Metacritic

Citation numbers refer to "Sources (Metacritic)". Several official pages were blocked (Fandom terms HTTP 402, press behind 403), so thresholds and the 15-critic minimum rest on Wikipedia and should be checked against Metacritic's live help page before any product decision.

### 4.1 Information presentation

- **Metascore box**: green = 75+ "generally favorable" [3]. Bands per Wikipedia [5]: games 90–100 universal acclaim, 75–89 generally favorable, 50–74 mixed or average, 20–49 generally unfavorable, 0–19 overwhelming dislike; films/TV/music use 81/61/40/20 cut-offs. Colour names per band not stated in the sources read.
- **Badges** [5]: Must-See (films, since June 2018) = Metascore 81+ and at least 15 critics; Must-Play (games, since September 2018) = 90+ and at least 15 industry reviews.
- **Critic reviews**: publication, score, excerpt, link to the original review (a link is required by the Fabric licensing terms [9]). Per-platform game pages exist; whether the Metascore is split per platform `[NOT VERIFIED]`.
- **User score**: 0–10 with one decimal [1][5].
- **Browse filters** (platform, genre, year, minimum Metascore, sort): described only by third-party scrapers [16] `[NOT VERIFIED]`.
- **2023 redesign under Fandom**: no evidence found; the 2023 change was moderation policy (4.3). The Fandom acquisition was announced 2022-10-03 [6][7].
- Mobile apps, "Best of" lists: `[NOT VERIFIED]`.

### 4.2 API and data access

- No public API is documented. Licensed access goes through Fabric Origin: developer approval, paid subscription to a Fabric Origin core solution, separate terms; no free-trial access [9][10].
- Fabric Origin terms for Metacritic data [9]: display allowed; ad sales or keyword bidding on the data need Fandom's written approval; sharing with third parties only as anonymous aggregates; no editing of items; scores and critic quotes must link to the source; fees waived when the usage rules are followed; quarterly usage reports; auto-renewing one-year term; cease use within five business days after termination. No clauses on caching or scraping in that document.
- Fandom terms of use: `/terms` and `/terms-of-use` returned HTTP 402 [11]; anti-scraping language appears only in secondary quotes `[NOT VERIFIED]`.
- Storefront integrations: Steam shows a Metascore where the developer links a Metacritic page, per community posts `[NOT VERIFIED]`; PlayStation Store: no source found. RSS or feeds: none found.

### 4.3 Rating calculation

- Metascore = weighted average of critic scores; weights reflect publication stature; the formula is not public [1][3][5].
- Weights: a 2013 Full Sail study reported six tiers (1.5, 1.25, 1.0, 0.75, 0.5, 0.25); Metacritic called it "wildly, wholly inaccurate" [4].
- Normalisation: reviews converted to 0–100 from the stated rating or by editorial judgement [5]. Letter-grade mapping (B=75, C=50, D=25 in blogs) `[NOT VERIFIED]`.
- Minimum critics for a Metascore: four, per a search snippet of the About page [2]. Must-See/Must-Play need 15 [5].
- User-score display minimum: one blog says 20+ ratings [17] `[NOT VERIFIED]`.
- Games older than one year are not accepted for new Metascores [3].
- Critic inclusion: no objective published criteria; audience, quality and publisher suggestions are considered [3].
- Review bombing: September 2011 user bans for "bombing" [5]. July 2020, after *The Last of Us Part II*: 36-hour wait before user reviews open, with the message "Please spend some time playing the game"; Metacritic described the change as data-driven [5][12].
- 2023 moderation (after *Horizon Forbidden West: Burning Shores* abuse): the team reviews each abuse report and removes violating reviews; Metacritic promised to be "evolving our processes and tools to introduce stricter moderation in the coming months" [8]. No concrete rules in the sources read.

### 4.4 User interaction

- Users can rate and review [1]. Minimum review length, lists, collections, "My Games", reporting: `[NOT VERIFIED]`.
- Users cannot edit metadata; licensees cannot either [9]. No ownership verification before reviewing [12].

### 4.5 Data freshness and sourcing

- Critics are added by editorial staff from a curated list, updated from publication suggestions [3]. Scores are not changed on request; reviews may be removed for Early Access or beta builds [3].
- Feed latency and release-date coverage: `[NOT VERIFIED]`.

### 4.6 Weaknesses and opportunities

- **Weaknesses**: opaque weighting [4][5]; disputed calculation; review bombing and the delay workaround [5][12]; reviews deleted by moderation without published rules [8]; no ownership check [12]; licensing restricts reuse [9]; opacity criticised in the industry since at least 2017 [3].
- **Opportunities for HexaScore**: publish the weighting model, normalisation table and minimum-count rules; show confidence and counts beside scores; cohort breakdown (verified-play vs all users); verified-play gating or weighting for user ratings; cross-media scores; explicit moderation policy with appeal.

### Sources (Metacritic)

1. <https://www.metacritic.com/about-metascores> (fetched), accessed 2026-10-28
2. <https://legacy.prod.metacritic.com/about-metascores> (search snippet; fetch refused), accessed 2026-10-28
3. <https://gamedeveloper.com/business/metacritic-101-for-indie-studios> (fetched; 2017-04-21), accessed 2026-10-28
4. <https://www.gamedeveloper.com/business/full-sail-study-attempts-to-shed-light-on-metacritic-s-weighting-system> (fetched; 2013-03-27), accessed 2026-10-28
5. <https://en.wikipedia.org/wiki/Metacritic> (fetched; secondary), accessed 2026-10-28
6. <https://www.hollywoodreporter.com/business/digital/tv-guide-metacritic-gamespot-acquisition-fandom-1235231819/> (cited via Wikipedia), accessed 2026-10-28
7. <https://www.aap.com.au/aapreleases/cision20221003ae90371> (Fandom release, 2022-10-03), accessed 2026-10-28
8. <https://www.ggrecon.com/articles/metacritic-changes-review-process-after-horizon-dlc-scandal/> (fetched; 2023-04-26), accessed 2026-10-28
9. <https://www.fabricdata.com/developer/metacritic-tos> (fetched; licensing terms), accessed 2026-10-28
10. <https://www.fabricdata.com/> (search result), accessed 2026-10-28
11. <https://www.fandom.com/terms> and <https://www.fandom.com/terms-of-use> (HTTP 402), accessed 2026-10-28
12. <https://knowyourmeme.com/news/metacritic-tweaks-user-review-system> (search snippet; July 2020), accessed 2026-10-28
13. <https://videogameschronicle.com/news/metacritic-pledges-stricter-moderation-after-abusive-horizon-forbidden-west-dlc-review-bombing> (HTTP 403), accessed 2026-10-28
14. <https://www.gamespot.com/articles/metacritic-will-delay-user-reviews-for-games/1100-6479886/> (HTTP 403), accessed 2026-10-28
15. <https://www.theverge.com/2020/7/17/21327946/metacritic-user-review-delay-36-hours-the-last-of-us-part-ii> (fetch refused; cited via Wikipedia), accessed 2026-10-28
16. <https://apify.com/fetch_cat/metacritic-reviews-scores-scraper> (third-party, unverified), accessed 2026-10-28
17. <https://tech-insider.org/?p=21079> (unverified 20-rating claim), accessed 2026-10-28

## 5. Rotten Tomatoes

Citation numbers refer to "Sources (Rotten Tomatoes)". Most facts come from Rotten Tomatoes' own FAQ, about, criteria and licensing pages fetched on 2026-10-28; the fetch tool returned summaries, so wording is paraphrased unless quoted. Two editorial articles carry a 2026 date that conflicts with 2024 content; they are not used as fact.

### 5.1 Information presentation

- **Title page**: Tomatometer (red tomato = Fresh, green splat = Rotten, grey = not enough reviews or unreleased) and Popcornmeter (red "Hot" at 60%+, green "Stale" below). Certified Fresh and Verified Hot are badges next to the scores [1][2][3][4].
- **2024 rebrand**: the press release of 2024-08-21 introduced Verified Hot and the Hot/Stale labels and describes enhanced Popcornmeter branding [5]; Wikipedia and NBCU coverage describe the audience score as renamed to Popcornmeter [6][7]. Treat as "renamed and branded in 2024" with a source discrepancy.
- Critics consensus, where-to-watch, cast, Top Critics filter, browse filters, editorial lists, mobile layout: `[NOT VERIFIED]` (not fetched).

### 5.2 API and data access

- The public developer API is effectively closed: the old key-based portal became invite-only by about 2021 and the Fandango-run portal required an application by 2022 [9]; exact closure date `[NOT VERIFIED]`. developer.fandango.com was unreachable (DNS error) [11].
- Official route: a Business Proposal Form for licensing or partnership; requests are reviewed before access; brand attribution is required wherever content is displayed [4][10].
- Terms prohibit logo use without written permission and require scores and reviews to link back to the Rotten Tomatoes title page [10].
- Pricing `[NOT VERIFIED]`: third-party sites cite about US$60,000 per year and a Fabric/Origin route; not confirmed on an official page.
- Feeds: none found. Streaming availability data source not confirmed; JustWatch links exist via TMDB, not via a Rotten Tomatoes statement [12].

### 5.3 Rating calculation

- **Tomatometer** = share of approved critic reviews classified as positive (binary). Fresh = 60% and up; Rotten = 59% and below [1][2].
- **Certified Fresh** (films): Tomatometer ≥75%, at least 80 reviews for wide release, 40 for limited release, 20 for TV seasons, and at least 5 Top Critic reviews; kept until the score falls below 70% [10][1].
- **Top Critics** cannot be applied for; selected by an internal team and an advisory board [2].
- **Average Rating** is computed separately from critic star ratings or letter grades on a 10-point scale, excluding unrated reviews [2].
- **Popcornmeter** = share of users rating a title 3.5/5 or higher; movie minimum 50 audience ratings or reviews [1][2].
- **Verified ratings**: users whose Fandango ticket purchase can be confirmed; started May 2019; US theatrical releases only [2][13].
- **Verified Hot**: verified Popcornmeter ≥90% with 500 verified ratings for wide release and 250 for limited; lost below 80% [10][4]. A Hollywood Reporter snippet gives different minimums (10/20/40 by box-office projection) [14]; the licensing page is treated as authoritative.
- **Anti-manipulation**: "Want to See" pre-release audience scores removed on 2019-02-25 [15]; verified ratings followed in May 2019 [13]. September 2023: Vulture reported that PR firm Bunker 15 paid obscure critics to lift *Ophelia* from 46% to 62%; Rotten Tomatoes delisted the film and said it does not tolerate manipulation [16, via JoBlo; Vulture itself not reachable]. The FAQ does not address review bombing directly; "Community Watch" titles require a written review from unverified users and may take longer to process [2].

### 5.4 User interaction

- User ratings 0.5–5 stars in 0.5 steps, written reviews, watchlist; verified users get a badge and the default audience score uses verified ratings only where available [1][2]. Star-scale and watchlist details `[NOT VERIFIED]`.
- Users cannot edit catalog data; corrections go through a form and staff check their sources [2].
- **Critic applications**: Tomatometer-approved individuals and outlets are approved case by case; applications open each March (next window 2027-03-01); written critics need two years of consistent output and a credible publication, or a site with 200,000 monthly visitors or 5,000 newsletter subscribers; video critics need 30,000 YouTube subscribers; reviews must focus substantially on the title and social posts do not qualify [3].

### 5.5 Data freshness and sourcing

- Reviews enter through approved critics and outlets; staff add titles and monitor sources; critics submit reviews via a form and can report missing reviews the same way [2][3].
- Scores appear only after the minimum review counts, so new titles show grey scores for the first days [2].
- Release calendar and time-to-first-review: `[NOT VERIFIED]`.

### 5.6 Weaknesses and opportunities

- **Weaknesses**: binary Fresh/Rotten loses nuance and the Average Rating is secondary [2]; the widened critic pool is gameable (Bunker 15 case) [16]; verification is Fandango-only, so it covers only US theatrical tickets from one vendor and creates a conflict of interest since Fandango owns the site [6][9]; verified ratings do not cover streaming, games or books; no games or books at all [2].
- **Opportunities for HexaScore**: publish a verified-activity methodology for all media (purchase, library or platform proof); show nuanced scores beside any binary summary; disclose critic approval criteria and excluded sources; publish how review bombing is handled.

### Sources (Rotten Tomatoes)

1. <https://www.rottentomatoes.com/help_desk> (fetched), accessed 2026-10-28
2. <https://www.rottentomatoes.com/faq> (fetched), accessed 2026-10-28
3. <https://www.rottentomatoes.com/critics/criteria> (fetched), accessed 2026-10-28
4. <https://www.rottentomatoes.com/about> (fetched), accessed 2026-10-28
5. <https://pressroom.versantmedia.com/rotten-tomatoes/press-releases/rotten-tomatoes-unveils-new-rotten-tomatoes-verified-hot-audience> (2024-08-21), accessed 2026-10-28
6. <https://en.wikipedia.org/wiki/Rotten_Tomatoes>, accessed 2026-10-28
7. <https://www.nbcuniversal.com/article/rotten-tomatoes-unveils-new-rotten-tomatoes-verified-hot-audience-badge> (search snippet), accessed 2026-10-28
8. <https://editorial.rottentomatoes.com/article/rotten-tomatoes-score-update> (date conflict; not cited as fact), accessed 2026-10-28
9. <https://www.rottentomatoes.com/help_desk/licensing> (fetched), accessed 2026-10-28
10. <https://editorial.rottentomatoes.com/article/audience-score-update/> (date conflict; content matches the 2024 release in part), accessed 2026-10-28
11. <https://developer.fandango.com> (unreachable, DNS error), accessed 2026-10-28
12. <https://www.justwatch.com> (search only), accessed 2026-10-28
13. <https://theplaylist.net/rotten-tomatoes-verified-audience-ratings-20190523/> (search snippet), accessed 2026-10-28
14. <https://www.hollywoodreporter.com/business/business-news/new-rotten-tomatoes-score-badge-verified-hot-1235974680/> (search snippet; secondary), accessed 2026-10-28
15. <https://www.engadget.com/2019-02-26-rotten-tomatoes-trolls-audience-scores.html> (search snippet), accessed 2026-10-28
16. <https://www.joblo.com/rotten-tomatoes-paying-for-reviews/> (secondary; Vulture original not reachable), accessed 2026-10-28

## 6. Goodreads

Citation numbers refer to "Sources (Goodreads)". Coverage is partial: help.goodreads.com, the Terms of Use and the Librarian Manual returned 404, 403 or timed out, so those points rely on secondary sources and are tagged `[NOT VERIFIED]`.

### 6.1 Information presentation

- **Book page** (2022+ redesign): a "Want to Read" button and a refreshed layout rolled out from mid-2022 [7][8]; users complained about the desktop layout in June 2023 (Goodreads status update, community). Star distribution, edition list, review sort/filter: `[NOT VERIFIED]`.
- **Goodreads Choice Awards** [3]: 20 nominees per category; nominees need an average rating of at least 3.50 at launch; one genre category per book; one book per series per category; two voting rounds; eligibility is a publication window (2026 cycle: 2025-11-12 to 2026-11-10).
- Listopia, series pages, Reading Challenge, mobile app specifics: `[NOT VERIFIED]`. The CSV export is not available in the app [9, third party].

### 6.2 API and data access

- Goodreads stopped issuing new developer keys on 2020-12-08 and said it "plans to retire these tools" [1][2]. Reports of key deactivation after 30 days of inactivity and of reduced fields around 2022 are single-source `[NOT VERIFIED]`.
- Replacement: none. The developer terms reportedly "don't allow storing of any Goodreads data" [2].
- Amazon Product Advertising API documents Amazon product reviews, not Goodreads data; no source says it covers Goodreads [11]. Not a Goodreads data path.
- Personal export: My Books > Import and export > Export Library on desktop; a CSV with an "Exclusive Shelf" column (`read`, `currently-reading`, `to-read`); format undocumented [9][10].
- Terms of Use not fetched (404); scraping policy `[NOT VERIFIED]`.

### 6.3 Rating calculation

- Average is a decimal from integer 1–5 stars; no official formula was found. Staff said averages are cached, recalculated nightly, and can drift during the day [13, community-sourced]. The old API returned whole-number averages at one point [14]; current behaviour `[NOT VERIFIED]`.
- Two-decimal display (e.g. 4.62) seen on list pages via snippet. No minimum rating count and no weighting documented; "simple unweighted mean" is the working assumption `[NOT VERIFIED]`.
- Pre-publication ratings are allowed [15]; source-declaration controls for pre-release copies are reported, undated `[NOT VERIFIED]`.
- **Review bombing**: Cait Corrain (*Crown of Starlight*) created multiple fake accounts to leave 1-star reviews on debut authors; the publisher cancelled the book [5][6].
- **Goodreads response** (2023-10-30 statement): it will "temporarily limit ratings being submitted for books during times when activity for that book is unusually high", a measure introduced earlier in 2023; reviews must not "deflate or inflate the rating of a book unfairly" [4][18]. Removal of existing reviews is reported but not confirmed by the fetched text.

### 6.4 User interaction

- Shelves: Want to Read, Currently Reading, Read, plus custom shelves (shelf names appear in the export) [9][10]. Ratings, reviews, Q&A, groups, quotes, friends: `[NOT VERIFIED]` beyond export columns.
- **Librarians** can edit book and author data, add covers and combine editions [12]. Combining and merging are distinct; combining is restricted to editions with the same primary author (forum posts). Combine was temporarily disabled in April 2025 and again from 22 December to 12 January [12, snippet].
- **Becoming a librarian**: the "50 books on shelves" rule is forum-sourced, not an official rule, and meeting it does not guarantee acceptance; a 2021 blog describes a 30-minute open-book quiz [16] `[NOT VERIFIED]`. The Librarian Manual was not accessible [19].
- **Author Program**: free; "Is this you?" link on the author page; approval typically within two business days per author blogs [17, snippet]; rejections reported without explanation.

### 6.5 Data freshness and sourcing

- Sources: Amazon catalogue, publisher/librarian entry, user-added books. Publisher ONIX feeds and ISBN handling `[NOT VERIFIED]`.
- Edition combining is manual and librarian-driven, so new editions can sit as separate records until a librarian combines them [12].
- Author-claim approvals are not instant [17].

### 6.6 Weaknesses and opportunities

- **Documented criticism**: no half stars; simple mean; review bombing with a reactive 2023 measure; pre-release ratings; dated UI; API shutdown that killed the third-party ecosystem; Amazon dependency. Non-English coverage `[NOT VERIFIED]`.
- **Opportunities for HexaScore**: documented rating method (half stars, weighted mean with minimum counts, pre-release windows); public, versioned API with a stable licence; full data export for users; moderation pipeline that does not depend on a volunteer librarian backlog. StoryGraph, Hardcover and Literal were not researched here.

### Sources (Goodreads)

1. <https://www.goodreads.com/api> (banner via search snippet; fetch returned the homepage), accessed 2026-10-28
2. <https://stephanieawilkinson.com/2020/12/10/yonderbook-and-goodreads/> (quotes the December 2020 notice), accessed 2026-10-28
3. <https://www.goodreads.com/choiceawards> (fetched), accessed 2026-10-28
4. <https://bookriot.com/goodreads-releases-statement-on-review-bombing/> (fetched; 2023-11-02), accessed 2026-10-28
5. <https://www.npr.org/2023/12/17/1219599404/goodreads-review-bombing-cait-corrain> (snippet; fetch timed out), accessed 2026-10-28
6. <https://www.thereviewgeek.com/caitcorrain-crownofstarlight-reviewbombing/> (HTTP 403), accessed 2026-10-28
7. <https://uxdesign.cc/a-look-at-the-new-book-page-design-on-goodreads-c7b248520b64> (November 2022), accessed 2026-10-28
8. <https://www.androidauthority.com/goodreads-android-revamped-design-598048/>, accessed 2026-10-28
9. <https://pickupreader.com/blog/how-to-export-goodreads-library-2026> (third party), accessed 2026-10-28
10. <https://github.com/JohnSmithDev/GRAnalysis>, accessed 2026-10-28
11. <https://affiliate-program.amazon.com/help/node/topic/GM6Z4B6SQMSG7D8U>, accessed 2026-10-28
12. <https://www.goodreads.com/group/show/220> (Librarians group), accessed 2026-10-28
13. <https://www.goodreads.com/topic/show/19185342-average-rating-only-for-single-authored-books>, accessed 2026-10-28
14. <https://www.goodreads.com/topic/show/17945919-recent-reviews-giving-integer-values-for-average-rating?page=1>, accessed 2026-10-28
15. <https://bookriot.com/goodreads-pre-publication-reviews/>, accessed 2026-10-28
16. <https://diarium.usal.es/pmgallardo/2021/04/01/how-to-become-a-goodreads-librarian/> (2021), accessed 2026-10-28
17. <https://goodreads.com/author/program> (snippet), accessed 2026-10-28
18. <https://www.publishersweekly.com/pw/by-topic/industrynews/publisher-news/article/93651-goodreads-asks-users-to-help-combat-review-bombing.html> (snippet), accessed 2026-10-28
19. <https://help.goodreads.com/s/topic/0TO1H000000gjH3WAI/librarian-manual> (not accessible), accessed 2026-10-28

## 7. Skoob

Citation numbers refer to "Sources (Skoob)". Correction to the research brief: Skoob is not owned by Americanas any more. Americanas bought it in September 2021 via IF Capital [8][9] and sold it to Skeelo (announced 2024-08-19 to 2024-08-26) during Americanas' judicial recovery [4][5]; the privacy policy names Skeelo as "Empresa titular do Skeelo App e do Skoob" and the legal entity SKOOB BOOKS 3D as "de titularidade da Skeelo" [1]. Price not disclosed [4][5]. Help centre, "Sobre", "Termos de uso" and "Cortesias" pages sit behind a login wall [14], so terms could not be read.

### 7.1 Information presentation

- **Book page** (sample ISBN 8573051876) [2]: cover left; title, publisher line, year, pages, ISBN-10, language; five-star rating shown as "4.2 / 181" (one decimal) with a percentage distribution per star; tabs Sinopse, Edições, Vídeos, Grupos, Resenhas, Leitores, Similares; shelf counts labelled "Leram", "Lendo", "Querem", "Relendo", "Abandonos", "Favoritos", "Desejados", "Avaliaram"; buttons Adicionar, Compartilhar, Editar; no ads visible on that snapshot; no author link. Counts differ between snapshots [2][3].
- The Groups tab shows a maintenance notice: "essa funcionalidade está em manutenção e em processo de análise interna" [3].
- "Top Mais" (100 most read/wanted) and "Similares" exist [10] `[NOT VERIFIED]` (Tecnoblog returned 403). Paginômetro and a yearly challenge (one medal per 250 pages) [10] `[NOT VERIFIED]`.
- **Mobile**: iOS app "Skoob - Organize suas leituras", iOS 16.4+, age 9+ [15, search result]. Android `[NOT VERIFIED]`.
- **Redesign**: in May 2025 Skeelo said "atualizar a plataforma é uma prioridade" with a responsive web version as the next step [6][7]. No 2026 news found.

### 7.2 API and data access

- No public API found. Terms could not be read [14]; the privacy policy has no clause on scraping or automated access [1]. A community Ruby gem exists as an unofficial client; not sanctioned, not relied on.
- Academic use is mostly qualitative with small samples (a 2022 mapping review found 15 papers) [13]; a paper claiming 45,114 reviews for sentiment analysis could not be read [12] `[NOT VERIFIED]`.
- **Data export**: the privacy policy grants portability ("solicitar a portabilidade dos seus Dados", clauses 4.3 iv and 8.1 v) [1]; a self-serve export tool `[NOT VERIFIED]`.
- Profile sharing between Skeelo and Skoob for service integration; payment data shared among group companies (3.1 xi–xii) [1].
- Partnerships: the 2021 rationale was network effect for Americanas and Submarino [8]; store links on book pages `[NOT VERIFIED]`.

### 7.3 Rating calculation

- Scale 0–5 stars, one decimal displayed [2].
- Minimum count, weighting, tie-breaking: not found in accessible sources; unknown.
- "Top Mais" built from reads/wants, method undocumented `[NOT VERIFIED]`.
- Anti-manipulation: none documented; the privacy policy mentions only account-breach monitoring (3.7) [1].

### 7.4 User interaction

- Shelves: Leram, Lendo, Querem, Relendo, Abandonos, Favoritos, Desejados, Tenho [2].
- Resenhas and comments; 9 million reviews and 11 million registered users as of May 2025 [6][7].
- **Cortesias**: listed in the site footer [2]; Skeelo says publishers use Skoob for market research and courtesy copies, which informs print runs [6]. Request and eligibility rules `[NOT VERIFIED]` (login wall).
- Troca/PLUS (book swap) and Pedir emprestado (lending) are described by Tecnoblog [10] `[NOT VERIFIED]`; not visible on the sampled page.
- Groups paused [3]. Gamification (Paginômetro, challenge) `[NOT VERIFIED]`.
- Book addition, moderation ("curadoria") and author verification: `[NOT VERIFIED]`.

### 7.5 Data freshness and sourcing

- Edition tab counts (0 or 1) suggest per-work grouping [2][3]; duplicate handling `[NOT VERIFIED]`.
- ISBN-based registration, cover sources and release speed: `[NOT VERIFIED]`.
- Traffic value: 66% of Skeelo downloads in the 2025 Halloween campaign came from Skoob [6].

### 7.6 Weaknesses and opportunities

- **Weaknesses**: ownership churn (Americanas 2021, Skeelo 2024) [4][8]; seller in judicial recovery since January 2023 [16]; stalled features (groups in maintenance, responsive web still pending in 2025) [3][7]; advertising cookies and Google Analytics [1]; no API; terms behind login.
- **Market**: 10–11 million registered users and 1.3 million monthly active users (May 2025) [6][7]; publisher relations through Cortesias are its strongest asset [6].
- **Opportunities for HexaScore (pt-BR)**: publisher-facing courtesy-copy and market-research tooling; a public API and user export; multi-media catalogue where Skoob is books-only; transparent rating formula; ISBN-based edition deduplication with Brazilian editions. Skoob is moving toward Skeelo's ebook/audiobook ecosystem, so the book-social niche in Brazil is contested, not abandoned.

### Sources (Skoob)

1. <https://www.skoob.com.br/privacy-policies> (fetched), accessed 2026-10-28
2. <https://www.skoob.com.br/book/49704> (fetched), accessed 2026-10-28
3. <https://www.skoob.com.br/book/109868/groups> (fetched), accessed 2026-10-28
4. <https://mercadoeconsumo.com.br/19/08/2024/noticias-varejo/skeelo-ecossistema-de-livros-anuncia-compra-do-skoob-da-americanas/>, accessed 2026-10-28
5. <https://www.meioemensagem.com.br/midia/americanas-vende-rede-social-literaria-para-skeelo>, accessed 2026-10-28
6. <https://www.mobiletime.com.br/noticias/20/05/2025/skeelo-skoob-moderniza/>, accessed 2026-10-28
7. <https://www.publishnews.com.br/materias/2025/05/23/apanhadao-da-semana-skeelo-segue-com-o-processo-de-modernizacao-do-skoob>, accessed 2026-10-28
8. <https://mercadoeconsumo.com.br/2021/09/15/americanas-s-a-adquire-a-skoob-maior-plataforma-digital-literaria-do-pais/> (search snippet), accessed 2026-10-28
9. <https://www.baguete.com.br/noticias/16/09/2021/americanas-compra-skoob> (search snippet), accessed 2026-10-28
10. <https://tecnoblog.net/499637/como-usar-o-skoob-rede-social-para-leitores/> (HTTP 403), accessed 2026-10-28
11. <https://tecnoblog.net/noticias/2021/09/16/americanas-compra-skoob-a-maior-rede-social-de-livros-do-brasil/> (HTTP 403), accessed 2026-10-28
12. <https://www.redalyc.org/journal/147/14775278006/14775278006.pdf> (unreadable), accessed 2026-10-28
13. <https://seer.uniacademia.edu.br/index.php/verboDeMinas/article/download/3336/2340> (search snippet), accessed 2026-10-28
14. <https://www.skoob.com.br/ajuda>, /cortesias, /quem-somos, /sobre (login walls), accessed 2026-10-28
15. <https://apps.apple.com/BR/app/id904670263> (search result), accessed 2026-10-28
16. <https://static.poder360.com.br/2023/01/americanas-fato-relevante-recuperacao-20jan32023.pdf> (search result), accessed 2026-10-28

## 8. Comparative matrix

Values come from sections 1–7; "n/d" = not documented in the sources reached; `?` = `[NOT VERIFIED]`. The last column is the HexaScore target as proposed for Phase 2 (owner decisions pending).

| Dimension | IMDb | TMDB | IGDB | Metacritic | Rotten Tomatoes | Goodreads | Skoob | HexaScore (proposed) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Media | film, TV (games in paid bulk only) | film, TV | games | games, film, TV, music | film, TV | books | books | games, films, series, books |
| User score scale | 1–10 integer | 0–10 (shown as %) | 0–100 | 0–10, one decimal | 0.5–5 stars | 1–5 stars, no halves | 0–5 stars | 0.5–10 stars in half steps × 6 categories (D-00) |
| Aggregate formula | weighted average, undisclosed; Top 250 Bayesian `m=25,000` published | n/d (`vote_average` plain?) | n/d | critic Metascore weighted, undisclosed; user score mean | binary % positive (Tomatometer), % ≥3.5 (Popcornmeter); rules published | plain mean? nightly recompute | plain mean? | published Bayesian shrinkage per type, versioned, with confidence |
| Minimum count to show | 5 ratings (display); 25,000 for Top 250 | n/d | n/d | 4 critics (Metascore); 15 for badges | 50 audience ratings (films); 80/40/20 reviews for Certified Fresh | none | n/d | 5 today; keep, plus confidence band |
| Anti-manipulation | undisclosed alternate weighting; "unusual voting activity" notice | n/d | n/d | 36 h wait (2020); manual moderation (2023) | pre-release scores removed (2019); Fandango-verified ratings; delisting for paid reviews (2023) | throttle during unusual activity (2023) | n/d | published rules: pre-release window, burst limiter, account-age and verified-activity weighting, public notice on affected items |
| Critic data | Metascore shown, not used in score | none | `aggregated_rating` from critics | core product | core product | none | none | none at launch (licensing); external scores display-only where licence allows |
| Public API | AWS Data Exchange only; US$150k/yr listed | yes, free non-commercial; commercial by written agreement; ~40 rps; 6-month cache cap | yes, free incl. commercial with partnership + attribution; 4 rps, 8 in flight; `limit` ≤ 500 | none public; Fabric Origin licence | closed; proposal form | retired 2020 | none | self-serve keys, published plans, OpenAPI 3.1, SDKs |
| Bulk / sync | daily TSV (non-commercial) | daily ID exports; `changes` (24 h default, 14 d max) | partner CSV dumps daily; webhooks create/update/delete; `updated_at`, `checksum` | none | none | none | none | daily ID export + `changes` feed + webhooks for paid plans |
| Data redistribution by third parties | prohibited without licence | prohibited commercially; no resale; no AI training | allowed to store and serve; attribution; agreement text unread | display only, link back, no sharing | display only, link back | n/a | n/a | per-field licence label in the API; own data under a permissive licence |
| Community edits | yes, staff review, 1–7 days | yes, direct edits, moderators remove unsupported content | yes, staff-validated | no | no (correction form) | librarians only (edit, combine) | "Editar" button present; process `?` | suggestions with states, 3/day/user, audit trail, SLA < 24 h |
| Personal tracking | watchlist, check-ins, lists | watchlist, favourites, lists, ratings | want to play / playing / played | "My Games"? | watchlist | shelves + custom shelves, challenge | 8 shelves, Paginômetro, challenge | status (want/in progress/completed/dropped), time-to-beat, lists; reading/playing goals later |
| Trending / popularity | MOVIEmeter from page views, weekly | lifetime popularity + daily/weekly trending; undocumented formula | PopScore primitives (IGDB lists, Steam, Twitch) exposed raw | n/d | n/d | n/d | "Top Mais" `?` | own primitives, 30-day decayed window, top 10 per type (D-02) |
| Entity pages | people, companies | people, companies, networks, collections | companies, franchises, platforms, engines | publications | critics, publications | authors, series | authors `?` | people, organisations with roles, franchises, platforms, genres (D-01) |
| Mobile app | yes `?` | yes `?` | `?` | `?` | `?` | yes (Android, iOS) | yes (iOS; Android `?`) | React Native, Android first |
| Locales | English UI; akas data | many | many | English | English | English | pt-BR | en, pt-BR, es |

## 9. Implications for HexaScore

Opinions of the architect, grounded in sections 1–8. Each item names the phase where it is decided or built.

### 9.1 Licensing is the first gate, not a detail (Phase 2 decision, Phase 3 ADR)

- **TMDB**: the API terms bar commercial use without a written agreement, cap caching at 6 months and bar resale (2.2). HexaScore ingests TMDB into a paid API today (`legacy/PLAN.md` 2026-10-10 decision). Options: (a) sign a commercial agreement with TMDB before launch; (b) keep TMDB-derived fields out of the paid API and out of long-term storage, which with a 6-month cap is operationally hostile to a catalogue; (c) drop TMDB and rely on Wikidata/Wikipedia for film/TV. Recommendation: pursue (a) in parallel and design the data model so every field carries its source, so (b) or (c) can be executed by configuration if (a) fails.
- **IGDB**: the FAQ allows commercial use with a partnership and visible attribution and encourages local storage (3.2). Email <partner@igdb.com> now; keep the answer in the repo (`docs/legal/`). Read the Twitch Developer Services Agreement verbatim before relying on the FAQ.
- **IMDb**: the datasets are non-commercial only; do not ingest them. Ratings from IMDb cannot be shown without a licence.
- **Metacritic and Rotten Tomatoes**: scores are licensed products (4.2, 5.2). The legacy `external_scores` table is front-end only, which is the right instinct, but even display needs a licence unless the value is obtained from a CC0 source (Wikidata review-score statements) and attributed as such. Decide per source in Phase 2; the safe default is to drop them from the UI at launch and add critic data later under licence.
- **Wikidata (CC0) and Wikipedia (CC BY-SA)** remain the spine. Open Library's data licence was not researched here and must be checked before Phase 3.
- Product consequence: a **per-field provenance and licence label** in the data model and the API (`source`, `licence`, `attribution`), so the API can expose only redistributable fields per plan. This is also a DX differentiator: no competitor tells developers what they may do with each field.

### 9.2 Publish the score (Phase 2 decision, scoring SDD)

- Keep six categories with half-star input (D-00) and a 0–100 aggregate. Publish the aggregate formula: Bayesian shrinkage toward the per-type prior with a documented `m` per type (IMDb publishes `m=25,000` for one chart; nobody publishes it for the main score), a documented minimum count (keep 5), and a confidence band (count-based) shown next to the score. Version the formula; changes are announced like API changes.
- Anti-manipulation as published rules, not secrets: no ratings before release (Rotten Tomatoes 2019 lesson); burst limiter per item when activity is anomalous (Goodreads 2023); weight by account age and by verified activity (status `completed` plus time-to-beat or reading progress as the HexaScore equivalent of Fandango's ticket); public "unusual activity" notice on the item page (IMDb). Rules live in the SDD and in a public help page.
- Separate "quality" from "interest": score vs trending (D-02), as TMDB separates `vote_average` from `popularity` and IGDB separates ratings from PopScore.

### 9.3 Trending and recommendations from own primitives (recommendation SDD)

- Copy IGDB's PopScore shape: store raw primitives (views, status changes, list adds, ratings) per item per day, then compute weighted, decayed trend scores per type. Expose both the top-10 lists and, for paid API plans, the primitives, which no film/TV/book competitor does.
- Recommendations start content-based (genres, creators, co-occurrence in lists and statuses); collaborative filtering waits for real users. Document the minimum data needed before collaborative filtering is turned on.

### 9.4 Community edits with a visible SLA (community-edits SDD)

- Suggestions, not direct edits (TMDB's direct-edit model needs a moderator army; IMDb's 1–7 day staff review frustrates). States: draft, submitted, in review, approved, rejected, reverted, with public status per suggestion (IMDb contribute tracker), one independent source required for factual changes (TMDB bible), contributor accuracy tracked over time (IMDb charter), and a target moderation SLA under 24 hours. Limit 3 per user per day (requirement 5) plus account-age and accuracy-based throttles.
- Every applied suggestion produces an audit event with before/after and a one-click revert.

### 9.5 Developer experience as a product (public API SDD, ADR)

- Self-serve keys, public price list, published rate limits with `RateLimit-*` headers (vukem provides GCRA and headers natively), OpenAPI 3.1 generated from typed handlers (vukem `api.Register`), SDKs for TypeScript and Python (already generated in legacy), and three freshness mechanisms borrowed from the best in class: daily ID export files (TMDB), a `changes` feed with a bounded window (TMDB), and webhooks with a secret header (IGDB).
- Write endpoints for first-party clients (mobile) are a separate, session- or token-scoped surface; the paid read API stays read-only at launch.

### 9.6 Entities and SEO (catalog SDD)

- First-class people, organisations with roles, franchises, platforms and genres (D-01), each with a stable slug, counts and a page in the sitemap. This is where IMDb and TMDB get most of their long-tail search traffic, and where Goodreads (authors, series) and IGDB (companies, franchises) are strong; Skoob has no author link on the book page.

### 9.7 Books and Brazil (catalog SDD, Phase 2 decision)

- Edition handling is the hard problem (Goodreads librarians combine editions by hand; Skoob shows 0–1 editions per work). Model work vs edition explicitly from day one, keyed by Wikidata work QID and ISBN-13 per edition, with Brazilian editions as a first-class case.
- Skoob's moat is publisher relations (Cortesias). A publisher-facing programme is a Phase 2 product question, not a Phase 3 spec.
- Cover coverage for books is 34% in production today; Open Library covers plus ISBN-based lookups are the first fix.

### 9.8 Mobile (mobile SDD)

- Every incumbent with a consumer brand has an app. The React Native app consumes the same typed API as the web (OpenAPI → generated TypeScript client), shares design tokens (star widget, category colours), and ships Android first.

### 9.9 Operational lessons

- Moderation latency and transparency are competitive features; budget for them.
- Do not depend on one source for any media type (TMDB 6-month cache cap, IGDB 4 rps, Goodreads API death). The ingestion SDD keeps Wikidata as the identity spine and treats every other source as enrichment that can be switched off per field.

## 10. Sources

Each competitor section ends with its own numbered source list (per-section numbering, all accessed 2026-10-28). Totals: IMDb 18, TMDB 12, IGDB 13, Metacritic 17, Rotten Tomatoes 16, Goodreads 19, Skoob 16. Pages that could not be read are listed with their HTTP status so a later pass with a browser can complete them; the priority list for that pass is: Twitch Developer Services Agreement, IGDB terms of service, TMDB API terms (verbatim check), Fandom terms of use, Goodreads terms and Librarian Manual, Skoob "Termos de uso" and "Cortesias".
