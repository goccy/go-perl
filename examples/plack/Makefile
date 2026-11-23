# The example needs no perl installed:
#
#   make setup   gperl mod download fetches the distributions cpanfile.snapshot
#                pins into the module directory (~/gperl/pkg/mod), builds them
#                once, and writes go.work
#   make run     go run .   (serves on http://localhost:8091)
#   make test    go test ./...
#   make server  the same application as a gperl-built server binary
GPERL := go run ../../cmd/gperl

.PHONY: setup run test server clean

setup:
	$(GPERL) mod download

run:
	go run .

test:
	go test ./...

server:
	GPERL_DEV_REPLACE=github.com/goccy/go-perl=$(abspath ../..) $(GPERL) psgi build -o plack-server app.psgi

clean:
	rm -f go.work go.work.sum plack plack-server
