module github.com/gosre/gosre-agent

go 1.26.2

require (
	github.com/gosre/gosre-sdk v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/gosre/gosre-sdk => ../gosre-sdk
