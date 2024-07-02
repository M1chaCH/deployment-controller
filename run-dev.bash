#!/bin/bash

export APP_HOST="localhost" # use "0.0.0.0 in prod
export APP_PORT=8080

export DB_USER="postgres"
export DB_PASS="postgres"

go run main.go