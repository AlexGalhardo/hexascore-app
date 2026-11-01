package requests

import "github.com/AlexGalhardo/vukem/z"

// ProfileDetails is the details section of /profile (SDD 03 section 6.1).
type ProfileDetails struct {
	Name   string `json:"name"   form:"name"`
	Locale string `json:"locale" form:"locale"`
}

// ProfileDetailsSchema validates ProfileDetails.
var ProfileDetailsSchema = z.Struct[ProfileDetails](z.Shape{
	"name":   z.String().Trim().Min(1).Max(100),
	"locale": z.Enum("en", "pt-BR", "es"),
})

// ProfilePassword changes or sets the password; CurrentPassword is checked
// only when the user already has a password.
type ProfilePassword struct {
	CurrentPassword string `json:"currentPassword" form:"currentPassword"`
	NewPassword     string `json:"newPassword"     form:"newPassword"`
}

// ProfilePasswordSchema validates ProfilePassword.
var ProfilePasswordSchema = z.Struct[ProfilePassword](z.Shape{
	"currentPassword": z.String().Max(256).Default(""),
	"newPassword":     PasswordSchema,
})

// ProfilePrivacy is the privacy section.
type ProfilePrivacy struct {
	ProfilePrivate        bool `json:"profilePrivate"        form:"profilePrivate"`
	PersonalizationOptOut bool `json:"personalizationOptOut" form:"personalizationOptOut"`
}

// ProfileNotifications is the email notices section.
type ProfileNotifications struct {
	CommunityMailOptOut bool `json:"communityMailOptOut" form:"communityMailOptOut"`
}

// ProfileNotificationsSchema validates ProfileNotifications.
var ProfileNotificationsSchema = z.Struct[ProfileNotifications](z.Shape{
	"communityMailOptOut": z.Bool().Default(false),
})

// ProfilePrivacySchema validates ProfilePrivacy.
var ProfilePrivacySchema = z.Struct[ProfilePrivacy](z.Shape{
	"profilePrivate":        z.Bool().Default(false),
	"personalizationOptOut": z.Bool().Default(false),
})

// ProfileFeed is the feed opt-in switch (SDD 14 FR-14-18).
type ProfileFeed struct {
	FeedOptIn bool `json:"feedOptIn" form:"feedOptIn"`
}

// ProfileFeedSchema validates ProfileFeed.
var ProfileFeedSchema = z.Struct[ProfileFeed](z.Shape{"feedOptIn": z.Bool().Default(false)})

// Review is a review of a rated item (SDD 14 §12.10).
type Review struct {
	Body    string `json:"body"    form:"body"`
	Spoiler bool   `json:"spoiler" form:"spoiler"`
}

// ReviewSchema validates Review; the database also checks 1 to 5,000 characters.
var ReviewSchema = z.Struct[Review](z.Shape{
	"body":    z.String().Trim().Min(1).Max(5000),
	"spoiler": z.Bool().Default(false),
})
