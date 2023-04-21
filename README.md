# Simple storage server

The possible solution for one of the interview task. 

## Problem 
Необходимо разработать сервис хранения файлов.
На сервер по HTTP PUT присылают файл, его надо разрезать на 5 равных частей и сохранить на 5 серверах хранения. При запросе HTTP GET нужно достать куски, склеить и отдать файл.

- Один сервер для HTTP-запросов
- Несколько серверов (>5) для хранения кусков файлов
- Реализовать тестовый модуль для сервиса, который обеспечит проверку его функционала
- Сервера для хранения могут добавляться в систему в любой момент, но не могут удаляться из системы
- Предусмотреть равномерное заполнение серверов хранения

## Solution
### Logical lever
- [storage-server](internal/storageserver/storageserver.go): keeps chunks on physical volum. When stoarage-server starts it interact with chunk-server and register itself. Storage-server has two api endpoint for uploadin and downloading chunks.
- [chunk-manager](internal/chunkmanager/chunkmanager.go): keeps information of chunks placement. It splits file into chunks. Chunks destributed between existed storage-servers.
- [api-server](internal/apiserver/apiserver.go): handle incoming client requests. It interacts with chunk-manager requesting chunks distribution map for the given file and directly interaction with storage-servers downloading/uploading chunks. Api-server also split/combine file into/from chunks.

### Service level
There is two servers:
- [storage-server](cmd/storage-server/main.go): exposes [storage-server](internal/storageserver/storageserver.go) API through HTTP via [http handler](internal/entrypoint/http/storageserver/handler.go)
- [api-server](cmd/api-server/main.go): exposes [chunk-manager](internal/chunkmanager/chunkmanager.go) and [api-server](internal/apiserver/apiserver.go) API through HTTP via [http handler](internal/entrypoint/http/apiserver/handler.go)

Interaction between components implemented via http-clients called [endpoints](internal/endpoint)
Since logical components do not depend on server and client layers  the interaction can be easily changed to another protocol (gRPC/protobuf or own designed solution). HTTP interaction chosen because of simplicity and easily for implementation in the given timeframe for that test solution.

## How to run
### Docker
Built docker images:
```
make docker-images
```
Run api-server and storage-servers:
```
make docker-run-api-server
make docker-run-storage-server n=1
make docker-run-storage-server n=2
make docker-run-storage-server n=3
make docker-run-storage-server n=4
make docker-run-storage-server n=5
```
### Golang
```
make run-api
make run-ss n=1
make run-ss n=2
make run-ss n=3
make run-ss n=4
make run-ss n=5
```

## How to test
Golang:
```
make deps
make test
```
Prepare test file:
```
make test-prepare
```
Upload test file:
```
make test-upload
```
Download test file:
```
make test-download
```

## Further development
- Compact chunk storage
  Chunks can be combined together into one big files at the storage-server level. Storage-server need to keep addition mapping information about chunk/file/offset. Helps to iresuse the load on storage-server file system.
- Establish heartbeat between storage-server and chunk-manager. 
  In case of connection lost in the given time chunk-manager can exclude storage-server from file distribution. 
- Chunk replication 
  Keeps several copy of each chunk. Helps to increase FTT level. 
- Healing/Redistributing
  In case of emergensy (storage-server failure) recreate one of the lost copies and redistribute remaining chunks.
