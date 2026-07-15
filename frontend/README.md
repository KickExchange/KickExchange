# frontend

Next.js app (App Router, TypeScript, Tailwind) for KickExchange. Calls trading-service's GraphQL API directly - no separate gateway.

## Run it

```
cd infra/docker
docker compose up --build
```

Or locally against the compose stack:

```
cd frontend
npm install
npm run dev
```

## Pages

- `/` - search box over `searchAssets`, shows only players already tradable on the platform.
- `/admin/add-player` - type a player name or transfermarkt id, preview via `previewPlayer`, confirm to call `addPlayer`. No auth yet - unprotected until auth lands.

## Configuration (env vars)

| Var | Default |
|---|---|
| `NEXT_PUBLIC_GRAPHQL_URL` | `http://localhost:8080/graphql` |
