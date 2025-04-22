#!/bin/bash

# Load environment variables from .env file
export $(cat .github/workflows/.secrets | xargs)

# Use psql to execute the SQL query
psql $DATABASE_URL -c "INSERT INTO users (api_token, name, email) VALUES ('$AUTHORIZATION', 'TEST', 'test@test.com');"