FROM postgres:15.3

ARG STARTUP_SCRIPTS_DIR="."

COPY $STARTUP_SCRIPTS_DIR/init.sql /docker-entrypoint-initdb.d/a_init.sql

EXPOSE 5432