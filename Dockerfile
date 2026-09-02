# Built by goreleaser (binary at context root) or:
#   make generate && make frontend && make service
#   cp service/faridoon-service ./faridoon-service && docker build -t faridoon .
FROM golang:1.25-alpine AS sqlmigrate
RUN go install github.com/rubenv/sql-migrate/sql-migrate@v1.8.1

FROM alpine:3.24
LABEL org.opencontainers.image.source=https://github.com/jamesread/Faridoon
RUN apk add --no-cache ca-certificates
COPY --from=sqlmigrate /go/bin/sql-migrate /usr/bin/sql-migrate
EXPOSE 8080
COPY faridoon-service /usr/bin/faridoon-service
COPY service/config.yaml /config/config.yaml
COPY frontend/dist /app/frontend
COPY database /var/faridoon/database
COPY docker-entrypoint.sh /usr/local/bin/faridoon-entrypoint.sh
RUN chmod +x /usr/local/bin/faridoon-entrypoint.sh /usr/bin/faridoon-service /usr/bin/sql-migrate
WORKDIR /app
ENV FARIDOON_CONFIG_FILE=/config/config.yaml
ENV FARIDOON_STATIC_DIR=/app/frontend
VOLUME /config
ENTRYPOINT ["/usr/local/bin/faridoon-entrypoint.sh"]
