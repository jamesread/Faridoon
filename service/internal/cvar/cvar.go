package cvar

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeBool   = "bool"
)

const (
	KeySiteTitle                = "site_title"
	KeyEnableVoting             = "enable_voting"
	KeyEnableRegistration       = "enable_registration"
	KeyEnableGuestAdd           = "enable_guest_add"
	KeyEnableSyntaxHighlighting = "enable_syntax_highlighting"
)

type Def struct {
	Key         string
	MainType    string
	ValueInt    int
	ValueString string
}

func Defaults(siteTitle string) []Def {
	return []Def{
		{Key: KeySiteTitle, MainType: TypeString, ValueString: siteTitle},
		{Key: KeyEnableVoting, MainType: TypeBool, ValueInt: 0},
		{Key: KeyEnableRegistration, MainType: TypeBool, ValueInt: 1},
		{Key: KeyEnableGuestAdd, MainType: TypeBool, ValueInt: 1},
		{Key: KeyEnableSyntaxHighlighting, MainType: TypeBool, ValueInt: 0},
	}
}
