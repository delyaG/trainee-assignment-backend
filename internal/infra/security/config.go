package security

type Config struct {
	JWTPrivateKey string `long:"jwt-private-key" env:"JWT_PRIVATE_KEY" description:"Path to JWT private key" required:"yes"`
}
