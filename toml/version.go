package toml

import "fmt"

type Version int

const (
	Version1 Version = 1
)

func (v Version) Validate() error {
	switch v {
	case Version1:
		return nil
	default:
		return fmt.Errorf("%d is an invalid version", v)
	}
}

func (v Version) IsSupported() bool {
	return v.Validate() != nil
}

func (v Version) String() string {
	return fmt.Sprintf("v%d", int(v))
}

func Current() Version {
	return Version1
}

func SupportedVersions() []Version {
	return []Version{Version1}
}
