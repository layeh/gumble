package gumble

// Version represents a Mumble client or server version.
type Version struct {
	version                uint32
	release, os, osVersion string
}

// Version returns the semantic version information as a single unsigned
// integer. Bits 0-15 are the major version, bits 16-23 are the minor version,
// and bits 24-31 are the patch version.
func (v *Version) Version() uint {
	return uint(v.version)
}

// Release returns the name of the client.
func (v *Version) Release() string {
	return v.release
}

// Os returns the operating system name and version.
func (v *Version) Os() (os, osVersion string) {
	return v.os, v.osVersion
}

// SemanticVersion returns the struct's semantic version components.
func (v *Version) SemanticVersion() (major, minor, patch uint) {
	major = uint(v.version>>16) & 0xFFFF
	minor = uint(v.version>>8) & 0xFF
	patch = uint(v.version) & 0xFF
	return
}

func (v *Version) setSemanticVersion(major, minor, patch uint) {
	v.version = uint32(major&0xFFFF)<<16 | uint32(minor&0xFF)<<8 | uint32(patch&0xFF)
}
