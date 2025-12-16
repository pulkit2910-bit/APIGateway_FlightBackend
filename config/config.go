package config

type Route struct {
	Prefix  string `yaml:"prefix"`
	Service string `yaml:"service"`
	Port    int    `yaml:"port"`
}

type Config struct {
	Routes []Route `yaml:"routes"`
}