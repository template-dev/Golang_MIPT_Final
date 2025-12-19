## Запуск PostgreSQL и Redis (Docker)
### PostgreSQL
```
docker run --name cashapp-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=cashapp -p 5432:5432 -d postgres:16-alpine
```

### Redis
```
docker run -p 6379:6379 --name cashcraft-redis -d redis:7-alpine
```

### Env
```
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
$env:REDIS_ADDR="localhost:6379"
$env:REDIS_DB="0"
```

### Goose up
```
goose -dir .\ledger\migrations postgres "$env:DATABASE_URL" up
```

### Запуск Gateway
```
cd gateway
go run .
```

### Примеры cURL
```
Invoke-RestMethod `
  -Method POST `
  -Uri http://localhost:8080/api/budgets `
  -ContentType "application/json" `
  -Body '{"category":"food","limit":5000}'
```

```
Invoke-RestMethod `
  -Method POST `
  -Uri http://localhost:8080/api/transactions `
  -ContentType "application/json" `
  -Body '{"amount":1200,"category":"food","description":"groceries","date":"2025-12-19T00:00:00Z"}'
```

```
curl -Method POST "http://localhost:8080/api/transactions/bulk?workers=4" `
  -ContentType "application/json" `
  -Body '[{"amount":1200,"category":"food","description":"a","date":"2025-12-19T00:00:00Z"},{"amount":0,"category":"food","description":"bad","date":"2025-12-19T00:00:00Z"}]'
```