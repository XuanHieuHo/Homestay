DB_URL=postgresql://root:secret@localhost:5432/homestay?sslmode=disable
postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root homestay

dropdb:
	docker exec -it postgres12 dropdb homestay

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

gen:
	docker run --rm -v "D:\Study\Homestay:/src" -w /src kjconroy/sqlc generate

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown gen