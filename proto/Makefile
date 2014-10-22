PROTOC-GEN-GO = $(GOPATH)/bin/protoc-gen-go

Mumble.pb.go: Mumble.proto
	protoc --plugin=protoc-gen-go=$(PROTOC-GEN-GO) --go_out=. $<

clean:
	rm -f Mumble.pb.go

.PHONY: clean
