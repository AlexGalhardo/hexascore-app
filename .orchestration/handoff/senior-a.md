# senior-a handoff
Delivered this round (all in review, REV rows prio 0): FIX-csp-inline-style (1797815), FIX-covers-route (a455a53), FIX-app-client-regen (7dfc80c), FIX-passkeys-register-e2e (see report).
Earlier in-review rows from the previous handoff unchanged. Stack `hexascore-qa-sa` stopped, db hexascore_qa_sa dropped.
Pitfalls: python text mode writes CRLF (fix with sed -i 's/\r$//'); backslashes in bash heredocs/sed get mangled, use Edit; golangci-lint refuses parallel runs (retry loop); lighthouse --output-path must be relative on Windows (not /tmp).
