build: 
	go build -o bin/nigarani
run: build
	./bin/nigarani
test: build
	go test -v ./...
