package cvar

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeBool   = "bool"
)

const (
	KeySiteTitle                = "site_title"
	KeyQuotesPerPage            = "quotes_per_page"
	KeyEnableVoting             = "enable_voting"
	KeyEnableRegistration       = "enable_registration"
	KeyEnableGuestAdd           = "enable_guest_add"
	KeyGuestAddRequireApproval  = "guest_add_require_approval"
	KeyEnableSyntaxHighlighting = "enable_syntax_highlighting"
	KeyEnablePwaPrompt          = "enable_pwa_prompt" // #nosec G101 #nosecret

	DefaultQuotesPerPage = 5
	MaxQuotesPerPage     = 127

	CategorySite     = "Site"
	CategoryFeatures = "Features"
	CategoryAccess   = "Access"
)

type Def struct {
	Key         string
	MainType    string
	Title       string
	Description string
	Category    string
	ValueString string
	Ordinal     int
	ValueInt    int
}

func Defaults(siteTitle string) []Def {
	return []Def{
		{
			Key: KeySiteTitle, MainType: TypeString, ValueString: siteTitle,
			Title: "Site title", Description: "Shown in the site header and browser tab.",
			Category: CategorySite, Ordinal: 10,
		},
		{
			Key: KeyQuotesPerPage, MainType: TypeInt, ValueInt: DefaultQuotesPerPage,
			Title: "Quotes per page", Description: "Number of quotes shown on each listing page (1–127).",
			Category: CategorySite, Ordinal: 20,
		},
		{
			Key: KeyEnableVoting, MainType: TypeBool, ValueInt: 0,
			Title: "Enable voting", Description: "Show vote controls and allow users to vote on quotes.",
			Category: CategoryFeatures, Ordinal: 30,
		},
		{
			Key: KeyEnableSyntaxHighlighting, MainType: TypeBool, ValueInt: 0,
			Title: "Enable syntax highlighting", Description: "Show the syntax highlighting field when adding or editing quotes.",
			Category: CategoryFeatures, Ordinal: 40,
		},
		{
			Key: KeyEnablePwaPrompt, MainType: TypeBool, ValueInt: 0,
			Title: "Show PWA install prompt", Description: "Show a banner inviting users to install Faridoon as an app when supported by the browser.",
			Category: CategoryFeatures, Ordinal: 45,
		},
		{
			Key: KeyEnableRegistration, MainType: TypeBool, ValueInt: 1,
			Title: "Enable registration", Description: "Allow new users to create accounts.",
			Category: CategoryAccess, Ordinal: 50,
		},
		{
			Key: KeyEnableGuestAdd, MainType: TypeBool, ValueInt: 1,
			Title: "Enable guest submissions", Description: "Allow logged-out users to submit quotes.",
			Category: CategoryAccess, Ordinal: 60,
		},
		{
			Key: KeyGuestAddRequireApproval, MainType: TypeBool, ValueInt: 1,
			Title:       "Require approval for guest submissions",
			Description: "When enabled, quotes submitted by guests wait for approval before they are published. Absolutely should not be disabled for untrusted networks or instances exposed on the public internet.",
			Category:    CategoryAccess, Ordinal: 70,
		},
	}
}
