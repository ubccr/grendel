package sqlstore

import "net/url"

// Config the config for Sqlstore.
type Config struct {
	Driver string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Driver: "sqlite3",
}

func configDefault(config ...Config) Config {
	return ConfigDefault
}

func (c Config) DataSourceName(filename string, rw bool) string {
	p := c.connectionParams()
	if rw {
		p.Set("mode", "rwc")
	} else {
		p.Set("mode", "ro")
	}
	return filename + "?" + p.Encode()
}

func (c Config) connectionParams() url.Values {
	p := url.Values{}
	p.Set("_txlock", "immediate")
	p.Set("_journal_mode", "WAL")
	p.Set("_foreign_keys", "true")

	return p
}
