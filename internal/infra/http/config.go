package http

type Config struct {
	Address         string   `short:"a" long:"address" env:"ADDRESS" description:"Service address" required:"yes"`
	AllowedOrigins  []string `long:"allowed-origins" env:"ALLOWED_ORIGINS" description:"Allowed origins to use CORS" env-delim:"," required:"yes"`
	JWTPrivateKey   string   `long:"jwt-private-key" env:"JWT_PRIVATE_KEY" description:"Path to JWT private key" required:"yes"`
	CookiePath      string   `long:"cookie-path" env:"COOKIE_PATH" description:"Cookie path" required:"yes"`
	CookieDomain    string   `long:"cookie-domain" env:"COOKIE_DOMAIN" description:"Cookie domain" required:"yes"`
	BaseFrontendURL string   `long:"base-frontend-url" env:"BASE_FRONTEND_URL" description:"Base frontend URL" required:"yes"`
}
