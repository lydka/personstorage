# personstorage

Small HTTP service for saving and loading `Person` records in SQLite.

## Build

```bash
make build
```

This produces the binary at `bin/personstorage`.

## Test

```bash
make test
```

Run only the end-to-end integration test:

```bash
make test-integration
```

## Run

```bash
./bin/personstorage
```

The server listens on `http://localhost:8080` and stores data in `data/app.db`.

If `8080` is already in use, run on another port:

```bash
LISTEN_ADDR=:18080 ./bin/personstorage
```

If you want a different SQLite file:

```bash
DATABASE_PATH=/tmp/personstorage.db ./bin/personstorage
```

## Docker

Build the image:

```bash
docker build -t personstorage .
```

Run the container and expose the HTTP server on `localhost:8080`:

```bash
docker run --rm -p 8080:8080 personstorage
```

If you want the SQLite database to persist outside the container:

```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/data:/app/data" \
  personstorage
```

Test the container with `curl`:

```bash
curl -i \
  -X POST http://localhost:8080/save \
  -H 'Content-Type: application/json' \
  -d '{
    "external_id":"docker-person-1",
    "name":"Ada Lovelace",
    "email":"ada.docker@example.com",
    "date_of_birth":"1815-12-10"
  }'

curl -i http://localhost:8080/docker-person-1
curl -i http://localhost:8080/missing-id
```

## Try It With curl

Create a person:

```bash
curl -i \
  -X POST http://localhost:18080/save \
  -H 'Content-Type: application/json' \
  -d '{
    "external_id":"123e4567-e89b-12d3-a456-426614174000",
    "name":"Ada Lovelace",
    "email":"ada@example.com",
    "date_of_birth":"1815-12-10"
  }'
```

Expected response:

```json
{"message":"Successfully saved"}
```

Fetch that person back:

```bash
curl -i http://localhost:18080/123e4567-e89b-12d3-a456-426614174000
```

Expected response:

```json
{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"Ada Lovelace","email":"ada@example.com","date_of_birth":"1815-12-10"}
```

Try a missing record:

```bash
curl -i http://localhost:18080/missing-id
```

Expected response:

```json
{"error":"Person not found"}
```
