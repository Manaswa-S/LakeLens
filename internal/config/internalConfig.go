package configs


type Paths struct {
	ResetPassHTMLPath string
}


type InternalConfig struct {
	BcryptPasswordCost int
	BcryptTokenCost    int

	Paths Paths
}

func InitInternalConfig() InternalConfig {
	return InternalConfig{
		BcryptPasswordCost: 12,
		BcryptTokenCost:    12,

		Paths: Paths{
			ResetPassHTMLPath: "/home/mnswa/zdev/go/projects/LakeLens/templates/emails/reset-pass.html",
		},
	}
}
