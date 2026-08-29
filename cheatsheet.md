# nstall Postgres v15 or late
sudo apt update
sudo apt install postgresql postgresql-contrib

# Update postgres password
sudo passwd postgres

# Start the Postgres server in the background
sudo service postgresql start

# Connect to the server.
sudo -u postgres psql

# Create a new database
CREATE DATABASE chirpy;

# Set the user password
ALTER USER postgres WITH PASSWORD 'postgres';

# Install Goose.
go install github.com/pressly/goose/v3/cmd/goose@latest

# Create a users migration in a new sql/schema directory
number_name.sql

# Run the up migration
goose postgres <connection_string> up

# Run th
goose postgres "postgres://wagslane:@localhost:5432/chirpy" up

# Install SQLC.
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# UUID Package
go get github.com/google/uuid

# Postgres Driver
go get github.com/lib/pq

# .env package
go get github.com/joho/godotenv

# Every http.Request has a context:
ctx := r.Context()

# Many database APIs accept a context.Context as their first argument. SQLC-generated methods are no exception:
user, err := cfg.db.CreateUser(ctx, params.Email)