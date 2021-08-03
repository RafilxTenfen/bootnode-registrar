build:
	go build -o bootnode-registrar *.go

run:
	go run main.go

compile:
	echo "Compiling for every OS and Platform"
	env GOTRACEBACK=none CGO_ENABLED=0 GOTRACEBACK=none GOOS=linux GOARCH=arm go build -trimpath -ldflags "-w -s" -o bin/bootnode-registrar-linux-arm *.go
	env GOTRACEBACK=none CGO_ENABLED=0 GOTRACEBACK=none GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-w -s" -o bin/bootnode-registrar-linux-arm64 *.go
	env GOTRACEBACK=none CGO_ENABLED=0 GOTRACEBACK=none GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-w -s" -o bin/bootnode-registrar-linux-amd64 *.go
	env GOTRACEBACK=none CGO_ENABLED=0 GOTRACEBACK=none GOOS=freebsd GOARCH=386 go build -trimpath -ldflags "-w -s" -o bin/bootnode-registrar-freebsd-386 *.go

all: build compile