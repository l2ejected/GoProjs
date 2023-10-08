build:
	go build -o bin/gobank

run: build
	./bin/gobank

test:
	go test -v ./...
	
up:
	docker run --name some-postgres -e POSTGRES_PASSWORD=gobank -p 5432:5432 -d postgres 
