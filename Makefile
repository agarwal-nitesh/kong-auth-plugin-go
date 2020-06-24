.PHONY: all

all: zkrull-auth.so

zkrull-auth.so: zkrull-auth.go
	go build -buildmode=plugin zkrull-auth.go
