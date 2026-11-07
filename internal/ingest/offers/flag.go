package offers

import "github.com/AlexGalhardo/vukem/flags"

// VerifyFlag is the kill switch of the verify job, which requests store pages (RFC-003 section 3). Default off until
// the owner enables it in production when ready. Env FLAG_OFFERS_VERIFY.
var VerifyFlag = flags.Bool("offers_verify", false).
	Describe("Daily check that offer links still answer (HEAD, 1 request per second per store)")
