BINARY_SERVER=vpn_server
BINARY_CLIENT=vpn_client

SRC_SERVER=./server/main.go
SRC_CLIENT=./client/main.go

.PHONY: all build clean build-linux deploy help

build:
	go build -o $(BINARY_CLIENT) $(SRC_CLIENT)
	go build -o $(BINARY_SERVER) $(SRC_SERVER)

clean:
	rm -f $(BINARY_CLIENT) $(BINARY_SERVER)

