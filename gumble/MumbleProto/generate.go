//go:generate wget -O Mumble.proto https://raw.githubusercontent.com/mumble-voip/mumble/master/src/Mumble.proto
//go:generate protoc --go_out=. Mumble.proto
//go:generate rm -f Mumble.proto
package MumbleProto
