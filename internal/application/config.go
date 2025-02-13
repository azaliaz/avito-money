package application

type Config struct {
	Name   string `env:"NAME" envDefault:"labels-api" yaml:"name"`
	Secret string `env:"SECRET" yaml:"secret"`
}
