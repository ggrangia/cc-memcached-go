build:
	go build -v -o cc-memcached-go cmd/cc-memcached-go/main.go

run:
	./cc-memcached-go -p 4000

br: build run 