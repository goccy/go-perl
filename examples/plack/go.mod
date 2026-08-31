module github.com/goccy/go-perl/examples/plack

go 1.25.0

require github.com/goccy/go-perl v0.0.0

require (
	github.com/ebitengine/purego v0.10.2 // indirect
	github.com/goccy/perlwasm2go v0.2.1 // indirect
)

replace github.com/goccy/go-perl => ../..

replace github.com/goccy/perlwasm2go => ../../../perlwasm2go
