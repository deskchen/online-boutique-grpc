module github.com/appnetorg/OnlineBoutique

go 1.22

replace github.com/appnetorg/OnlineBoutique/services => ./services

require (
	github.com/go-playground/validator/v10 v10.24.0
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.7.0
	google.golang.org/grpc v1.69.2
	google.golang.org/protobuf v1.36.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
)
