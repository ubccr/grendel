# Migration commands

```bash
# cli install
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# add a migration
migrate create -dir internal/store/migrations/sql -ext .sql  <name>

# change db versions
migrate -path internal/store/migrations/sql -database sqlite3://grendel.db up 1
migrate -path internal/store/migrations/sql -database sqlite3://grendel.db down 1

# force version on failed migration
migrate -path internal/store/migrations/sql -database sqlite3://grendel.db force <version>
```
