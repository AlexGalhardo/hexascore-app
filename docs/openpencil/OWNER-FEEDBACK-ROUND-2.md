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
