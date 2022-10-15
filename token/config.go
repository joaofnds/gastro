package token

type Config struct {
	PublicKey  string `mapstructure:"public_key"`
	PrivateKey string `mapstructure:"private_key"`
}
