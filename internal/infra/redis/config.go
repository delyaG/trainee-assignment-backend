package redis

type Config struct {
	Addr     string `long:"addr" env:"ADDR" description:"Redis address (host:port)" required:"yes"`
	Password string `long:"password" env:"PASSWORD" description:"Redis password" required:"yes"`
	DB       int    `long:"db" env:"DB" description:"Redis database number" required:"yes"`
	Dev      bool   `long:"dev" env:"DEV" description:"Enables 123456 code to use at development environments"`
}
