# Owner feedback, round 2 (2026-10-31)

The owner chose **v1 Editorial**. v2 Noir and v3 Paper are frozen: no more work on them.

## Changes to v1

1. Bring from v3: the score block with its explanation ("High score · from N ratings · six categories, one formula"), genre and platform chips (not stacked on the hero title), and the "More games like this" strip with covers.
2. Orchestrator fixes:
   - scores in the sans font with tabular figures; Playfair only for headings;
   - category bars must never overlap the number: bar and value in separate columns;
   - primary action in brand purple, not red (red means a low score).
3. Dark theme: a dark version of every v1 frame, AA contrast checked.
4. Footer on mobile: links stacked vertically and centred.
5. Navbar search: a results dropdown with the first 8 matches. Each row: cover on the left (2:3), title and release year on the right, the item type below the title, and the final average score at the far right ("–" under 5 ratings). Show empty, loading and "see all results" states, desktop and mobile.
6. Rating widget for the signed-in user: 10 yellow stars per category, half stars allowed (each half star = 0.5, so 10 stars = 10.0). The numeric value always prints next to the stars; keyboard operable (radio group, arrow keys); targets at least 44 px (the half-star hit areas may sit inside a 44 px row).
7. Mobile navbar: the four main categories (Games, Movies, TV Series, Books) spread with `justify-content: space-between` across the full width.
8. Theme toggle icon: moon in light mode, sun in dark mode.
9. Desktop alignment: every desktop frame uses the same content container as the navbar (same max width and side gutters). Breadcrumbs, hero, sections, side panels and footer all start and end on the navbar edges; nothing runs wider or sits offset. Check every 1440 frame, light and dark.
10. Home (all breakpoints, light and dark): remove the intro hero text. The page starts directly with the Games section: 3 large cards with the trending games of the last 7 days, then a grid of other games below them. Movies, TV Series and Books follow in that order, each with the same pattern (3 trending cards, then other titles). Each section has a heading and a "see all" link to its listing.
11. Category listing pages (Games, Movies, TV Series, Books): remove the "Section X" eyebrow and the subtitle under the category title; the title stands alone.
12. Buttons: "Reset" in the filters becomes a real button (same secondary style as the others, not a text link). Every button in v1 gets depth: a subtle shadow at rest and a small pressed state (moves down 1 px, shadow shrinks), plus hover and visible focus. Make it one button component with these states in light and dark, and use it on every page. The pressed motion needs a reduced-motion alternative (no movement, only the shadow change).
13. Category listing pages: numbered pagination (previous, page numbers, next) both above and below the results, with the filters at the top. Show page 1, a middle page and the last page; on mobile the pagination collapses to previous / "page N of M" / next.
14. Language: a dropdown button in the navbar only (EN, PT, ES; current one marked), on every page. The navbar is fixed and stays visible while scrolling (show a scrolled state with a bottom border or shadow). Remove the language links from the footer.
15. Social sign-in: a "Continue with Google" button on the login and signup pages (and their mobile frames), above the email form with an "or" divider. Follow Google's sign-in branding guidelines (official G mark, light and dark variants, wording "Sign in with Google" / "Sign up with Google", translated); this is the one allowed exception to the no-third-party-logos rule, note it in the README. Same depth and pressed states as item 12.
