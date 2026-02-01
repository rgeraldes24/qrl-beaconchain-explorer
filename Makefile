GITCOMMIT=`git describe --always`
VERSION=`git describe --always --tags`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=github.com/theQRL/qrl-beaconchain-explorer
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE} -s -w"

all: explorer stats frontend-data-updater el-indexer rewards-exporter signatures misc

lint:
	echo 
	golint ./...

test:
	go test ./...

explorer:
	rm -rf bin/
	mkdir -p bin/
	go run cmd/bundle/main.go
	go build --ldflags=${LDFLAGS} -o bin/explorer cmd/explorer/main.go

stats:
	go build --ldflags=${LDFLAGS} -o bin/statistics cmd/statistics/main.go

frontend-data-updater:
	go build --ldflags=${LDFLAGS} -o bin/frontend-data-updater cmd/frontend-data-updater/main.go

rewards-exporter:
	go build --ldflags=${LDFLAGS} -o bin/rewards-exporter cmd/rewards-exporter/main.go

el-indexer:
	go build --ldflags=${LDFLAGS} -o bin/el-indexer cmd/el-indexer/main.go

signatures:
	go build --ldflags=${LDFLAGS} -o bin/signatures cmd/signatures/main.go

misc:
	go build --ldflags=${LDFLAGS} -o bin/misc cmd/misc/main.go
