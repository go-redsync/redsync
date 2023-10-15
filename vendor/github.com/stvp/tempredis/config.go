package tempredis

// Config is a key-value map of Redis config settings.
type Config map[string]string

func (c Config) Socket() string {
	return c["unixsocket"]
}
