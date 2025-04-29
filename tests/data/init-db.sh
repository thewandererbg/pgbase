#!/bin/bash
set -e

echo "Init script started..."

# Wait until PostgreSQL is ready
until pg_isready -U "$POSTGRES_USER"; do
  echo "Waiting for postgres to be ready..."
  sleep 1
done

# Check if pbtest database exists
DB_EXISTS=$(psql -U "$POSTGRES_USER" -tAc "SELECT 1 FROM pg_database WHERE datname='pbtest'")

if [ "$DB_EXISTS" != "1" ]; then
  echo "Creating pbtest database..."
  createdb -U "$POSTGRES_USER" pbtest
  createdb -U "$POSTGRES_USER" pbdata
  createdb -U "$POSTGRES_USER" pbaux
else
  echo "Database pbtest already exists. Skipping creation."
fi

# Create backups folders if it is not available

# Restore the dump into pbtest
if [ -s /backups/pbtest.dump ]; then
  echo "Restoring pbtest.dump into pbtest..."
  pg_restore -U "$POSTGRES_USER" -d pbtest /backups/pbtest.dump
else
  echo "pbtest.dump not found or empty. Skipping restore."
fi

# # Restore the dump into pbtest
# if [ -s /backups/pbtest.sql ]; then
#   echo "Restoring pbtest.sql into pbtest..."
#   psql -U "$POSTGRES_USER" -d pbtest -f /backups/pbtest.sql
# else
#   echo "pbtest.sql not found or empty. Skipping restore."
# fi


echo "Init script finished."
