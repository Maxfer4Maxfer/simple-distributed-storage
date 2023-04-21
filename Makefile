deps:
	$(info Installing binary dependencies...)
	go install github.com/golang/mock/mockgen@latest
	go install gotest.tools/gotestsum@latest

run-ss:
	mkdir -p data/ss$(n)
	go run cmd/storage-server/main.go \
        --address 0.0.0.0:900$(n) \
		--chunk-manager 0.0.0.0:9000 \
		--data-directory data/ss$(n) \
		--registration-retry-timeout 4

run-api:
	go run cmd/api-server/main.go \
        --address 0.0.0.0:9000 \
		--max-chunk-size-bytes 262144 \
		--erasure-coding-fraction 2 

docker-image-apiserver:
	docker build --tag apiserver -f deploy/apiserver.Dockerfile .

docker-image-storageserver:
	docker build --tag storageserver -f deploy/storageserver.Dockerfile .

docker-images: docker-image-apiserver docker-image-storageserver

docker-run-api-server:
	docker run --net host apiserver:latest \
		--address 0.0.0.0:9000 \
		--max-chunk-size-bytes 9223372036854775807 \
		--erasure-coding-fraction 5

docker-run-storage-server:
	docker run --net host storageserver:latest \
		--address 0.0.0.0:900$(n) \
		--chunk-manager 0.0.0.0:9000 \
		--data-directory data \
		--registration-retry-timeout 4

test-prepare:
	mkdir data
	curl https://www.9minecraft.net/wp-content/uploads/2019/03/Simple-Storage-Network-mod-for-minecraft-logo.png --output data/simple-storage-network.png

test-upload:
	curl -X PUT -F file='@data/simple-storage-network.png' http://127.0.0.1:9000
	ls -R data

test-upload-ss:
	curl -X PUT -F chunk='@data/simple-storage-network.png' http://0.0.0.0:9001

test-download:
	curl -X GET --output simple-storage-network.png http://127.0.0.1:9000/?id=simple-storage-network.png
	diff simple-storage-network.png data/simple-storage-network.png

mock:
	mockgen -source=internal/apiserver/apiserver.go -destination=tests/mock/apiserver_mock.go -package=mock

test:
	gotestsum --format=testname -- -race ./... 

test-simple:
	go test -race ./... 
