# infra

## docker/

`docker-compose.yml` brings up backend services under the `kickexchange` Compose project, all joined to the `exchange-network` bridge network so services can reach each other by service name (e.g. `matching-engine:9000`).

```
cd infra/docker
docker compose up --build      # foreground
docker compose up -d --build   # detached
docker compose down            # stop and remove
```

Add new services to `docker-compose.yml` under the same `exchange-network` as they're built.
