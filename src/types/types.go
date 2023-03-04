package types

// ConnectionType can be Active or Passive
type ConnectionType string

type File struct {
	Name      string
	Extension string

	Bytes []byte
}
