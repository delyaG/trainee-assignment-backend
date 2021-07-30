package email

type Config struct {
	Host     string `long:"host" env:"HOST" description:"SMTP host" required:"yes"`
	Port     int    `long:"port" env:"PORT" description:"SMTP port" required:"yes"`
	Username string `long:"username" env:"USERNAME" description:"SMTP username" required:"yes"`
	Password string `long:"password" env:"PASSWORD" description:"SMTP password"`

	BaseBackendURL  string `long:"base-backend-url" env:"BASE_BACKEND_URL" description:"Base backend URL" required:"yes"`
}
