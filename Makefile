clean:
	rm ./bin/*

testing:
	go test ./src/.

run:
	go run ./src/.

build:
	go build -o ./bin/mercurio ./src/*.go
