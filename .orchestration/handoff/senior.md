# senior handoff
Delivered (REVs queued): P1.H.4.1b, P1.I.2.1, ADOPT-VF-024/035/040/046, P1.I.2.2, P1.I.2.3 (stacked on I.2.2). Queue: remaining 4-* tasks (P1.H.3.2, P1.I.2.4, P1.I.3.1, P1.I.5.1, P1.I.4.1).
Pitfalls: e2e needs APP_KEY in shell, shared compose project hexascore (container hexascore-app-1, port 3000); docker build ~3-5 min, run in background; repeated e2e runs trip per-item rating burst limiter (update items set ratings_limited_until=null); backslashes in sed/node -e one-liners get eaten, use Edit/heredoc files; page props must not be named seo (shadows shared prop).
