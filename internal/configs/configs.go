package configs

import (
	"os"
	"trainee-assignment-backend/internal/infra/email"
	"trainee-assignment-backend/internal/infra/http"
	"trainee-assignment-backend/internal/infra/postgres"
	"trainee-assignment-backend/internal/infra/redis"
	"trainee-assignment-backend/internal/infra/security"
	"trainee-assignment-backend/internal/infra/sms"
	"trainee-assignment-backend/pkg/logging"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Logger   *logging.Config  `group:"Logger args" namespace:"logger" env-namespace:"TRAINEE_ASSIGNMENT_LOGGER"`
	Postgres *postgres.Config `group:"Postgres args" namespace:"postgres" env-namespace:"TRAINEE_ASSIGNMENT_POSTGRES"`
	HTTP     *http.Config     `group:"HTTP args" namespace:"http" env-namespace:"TRAINEE_ASSIGNMENT_HTTP"`
	Security *security.Config `group:"Security args" namespace:"security" env-namespace:"TRAINEE_ASSIGNMENT_SECURITY"`
	Redis    *redis.Config    `group:"Redis args" namespace:"redis" env-namespace:"TRAINEE_ASSIGNMENT_REDIS"`
	Email    *email.Config    `group:"Email args" namespace:"email" env-namespace:"TRAINEE_ASSIGNMENT_EMAIL"`
	SMS      *sms.Config      `group:"SMS args" namespace:"sms" env-namespace:"TRAINEE_ASSIGNMENT_SMS"`
}

func Parse() (*Config, error) {
	var config Config
	p := flags.NewParser(&config, flags.HelpFlag|flags.PassDoubleDash)

	_, err := p.ParseArgs(os.Args)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
