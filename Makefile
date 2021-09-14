.PHONY: gen clean server client test cert

.PHONY: gen
gen:
	protoc --proto_path=proto proto/*.proto --plugin=EXECUTABLE --go_out=. 

.PHONY: bufgen
bufgen:
	@buf generate \
        --path proto/keyboard_message.proto \
		--path proto/laptop_message.proto \
		--path proto/laptop_service.proto \
		--path proto/memory_message.proto \
		--path proto/processor_message.proto \
		--path proto/screen_message.proto \
		--path proto/storage_message.proto \
		--path proto/filter_message.proto \
		--path proto/auth_service.proto			

.PHONY: clean
clean:
	rm pb/*.go

.PHONY: cert
cert:
	cd cert; ./gen.sh; cd ..

.PHONY: server
server:
	go run cmd/server/main.go -port 8080

.PHONY: client
client:
	go run cmd/client/main.go -address 127.0.0.1:8080

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: dep-install
dep-install:
	@go get -u \
		google.golang.org/protobuf/cmd/protoc-gen-go \	
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/bufbuild/buf/cmd/buf \
		github.com/mgechev/revive