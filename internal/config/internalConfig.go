package configs

type InternalConfig struct {
	BcryptPasswordCost int
}

func InitInternalConfig() InternalConfig {
	return InternalConfig{
		BcryptPasswordCost: 12,
	}
}