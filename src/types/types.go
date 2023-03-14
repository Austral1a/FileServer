package types

// ConnectionType can be Active or Passive
type ConnectionType string

type File struct {
	// Name can be "filename.txt"
	Name string
	Data []byte
}
