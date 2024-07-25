module github.com/ohanan/LambdaSha/pkg/plugins

go 1.22.0

replace github.com/ohanan/LambdaSha/pkg/lsha => ../lsha

require (
	github.com/ohanan/LambdaSha v0.0.0-20240722123328-34dc99ab4edf
	github.com/ohanan/LambdaSha/pkg/lsha v0.0.0
)
