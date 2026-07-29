# Build: make generate && make frontend && make service
# Then: docker build -t faridoon .
FROM alpine:3.20
RUN apk add --no-cache ca-certificates sql-migrate
EXPOSE 8080
COPY service/faridoon-service /usr/bin/faridoon-service
COPY service/config.yaml /config/config.yaml
COPY frontend/dist /app/frontend
COPY database /var/faridoon/database
COPY docker-entrypoint.sh /usr/local/bin/faridoon-entrypoint.sh
RUN chmod +x /usr/local/bin/faridoon-entrypoint.sh /usr/bin/faridoon-service
WORKDIR /app
ENV FARIDOON_CONFIG_FILE=/config/config.yaml
ENV FARIDOON_STATIC_DIR=/app/frontend
VOLUME /config
ENTRYPOINT ["/usr/local/bin/faridoon-entrypoint.sh"]
