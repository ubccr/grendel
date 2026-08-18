package sqlstore

import (
	"fmt"
	"net/url"
)

// Config the config for Sqlstore.
type Config struct {
	Driver string

	// MaxReadConns caps the read-only connection pool. Zero falls back to database.max_read_conns, then to defaultMaxReadConns.
	MaxReadConns int
}

func (c Config) DataSourceName(filename string, rw bool) string {
	p := c.connectionParams()
	if rw {
		p.Set("mode", "rwc")
	} else {
		p.Set("mode", "ro")
	}
	return fmt.Sprintf("file:%s?%s", filename, p.Encode())
}

func (c Config) connectionParams() url.Values {
	p := url.Values{}
	p.Set("_txlock", "immediate")
	p.Set("_journal_mode", "WAL")
	p.Set("_foreign_keys", "true")

	return p
}
