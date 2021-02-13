clean:
	rm ./bin/*

run:
	go run ./src/*.go

build:
	go build -o ./bin/mercurio ./src/*.go
