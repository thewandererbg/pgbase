# pgbase

pgbase is a fork of [PocketBase](https://pocketbase.io), using PostgreSQL instead of SQLite.

Everything else — features, API, Admin UI, documentation — works the same as PocketBase.
The only difference: PostgreSQL powers the backend.

Based on [PocketBase](https://pocketbase.io) version 0.27.1

## Quickstart

```bash
git clone git@github.com:thewandererbg/pgbase.git
cd examples/base
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o pgbase
./pgbase serve
```

Set your database connection:

```bash
export PB_DATA_URI="postgres://user:pass@localhost:5432/pbdata?sslmode=disable&default_query_exec_mode=simple_protocol"
export PB_AUX_URI="postgres://user:pass@localhost:5432/pbaux?sslmode=disable&default_query_exec_mode=simple_protocol"
```

## Docker

```bash
git clone git@github.com:thewandererbg/pgbase.git
cd docker
cp .env.sample .env
docker compose up -d
```

Access the Admin UI at http://localhost:8090/_/.

## Docs

Use the [PocketBase documentation](https://pocketbase.io/docs/) — just remember you're running on PostgreSQL.

### Differences with PocketBase

pgbase follows PocketBase behavior and APIs as closely as possible, with the following differences:

- Uses PostgreSQL instead of SQLite for all data storage
- Backup feature is not supported
- Cascade delete behavior differs from PocketBase
- Optional multi-instance support via PostgreSQL `LISTEN / NOTIFY`

#### Multi-instance support

pgbase can run multiple instances connected to the same PostgreSQL database.

When `IsPubSubEnabled` config is enabled:
- Collection schema changes are propagated to all instances
- Settings changes are propagated to all instances
- Each instance reloads its local cache automatically

Limitations:
- Realtime API events are **not propagated** across instances

When `IsPubSubEnabled` config is disabled:
- All cache reloads are local to the current instance only


## License

MIT. Forked from [PocketBase](https://pocketbase.io).
