module github.com/gosre/gosre-agent

go 1.25.4

require (
	github.com/gosre/gosre-sdk v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.1
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/gosre/gosre-sdk => ../gosre-sdk
