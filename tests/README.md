
```
docker exec pb_db pg_dump -U postgres -d pbtest -Fc -f pbtest.dump
docker exec pb_db pg_restore -U postgres -d pbtest pbtest.dump

docker exec pb_db pg_dump -U postgres -d pbtest -f pbtest.sql
docker exec pb_db psql -U postgres -d pbtest -f pbtest.sql
```
