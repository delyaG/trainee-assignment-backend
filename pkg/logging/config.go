package logging

// Config contains logger configuration
type Config struct {
	Level string `short:"l" long:"level" env:"LEVEL" description:"Logger level" required:"yes" default:"error"`
}
