# pgbase

pgbase is a fork of [PocketBase](https://pocketbase.io), using PostgreSQL instead of SQLite.

Everything else — features, API, Admin UI, documentation — works the same as PocketBase.
The only difference: PostgreSQL powers the backend.

Based on [PocketBase](https://pocketbase.io) version 0.27.1

## Quickstart

```bash
git clone https://github.com/thewandererbg/pocketbase.git
cd examples/base
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o pocketbase
./pocketbase serve
```

Set your database connection:

```bash
export PB_DATA_URI="postgres://user:pass@localhost:5432/pbdata?sslmode=disable"
export PB_AUX_URI="postgres://user:pass@localhost:5432/pbaux?sslmode=disable"
```

Access the Admin UI at http://localhost:8090/_/.

## Docs

Use the [PocketBase documentation](https://pocketbase.io/docs/) — just remember you're running on PostgreSQL.

## License

MIT. Forked from [PocketBase](https://pocketbase.io).
