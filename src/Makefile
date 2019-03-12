all: node clock bb

clean:
	@rm -rf *.exe

node:
	go build -o node.exe ./cmd/node.go

clock:
	go build -o clock.exe ./cmd/clock.go

bb:
	go build -o bb.exe ./cmd/bb.go
