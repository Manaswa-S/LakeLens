package configs

type CoreCfg struct {
	BcryptPasswordCost int
	BcryptTokenCost    int

	EPassAuth   string
	GoogleOAuth string
}

func InitCoreCfg() CoreCfg {
	return CoreCfg{
		BcryptPasswordCost: 12,
		BcryptTokenCost:    12,

		EPassAuth:   "epass",
		GoogleOAuth: "goauth",
	}
}
