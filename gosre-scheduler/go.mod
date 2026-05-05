module github.com/viniciushammett/gosre/gosre-scheduler

go 1.26.2

require (
	github.com/viniciushammett/gosre/gosre-sdk v0.0.0-00010101000000-000000000000
	github.com/nats-io/nats.go v1.51.0
	github.com/redis/go-redis/v9 v9.18.0
	github.com/viniciushammett/gosre/gosre-events v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace (
	github.com/viniciushammett/gosre/gosre-sdk => ../gosre-sdk
	github.com/viniciushammett/gosre/gosre-events => ../gosre-events
)
