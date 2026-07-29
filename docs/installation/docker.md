# Install Faridoon with Docker

The container image for Faridoon can be found on GitHub Container Registry:

```bash
docker pull ghcr.io/jamesread/faridoon:latest
```

Faridoon container images are built for **amd64** and **arm64**.

```bash
docker run -it --name faridoon -p 8080:8080 \
  -e DB_HOST=mysql -e DB_PASS=hunter2 -e DB_USER=faridoon -e DB_NAME=faridoon \
  -v faridoon-config:/config \
  ghcr.io/jamesread/faridoon:latest
```

Mount `/config` and place a `config.yaml` there (see Configuration). Consider [Docker Compose](docker-compose.md) instead.
