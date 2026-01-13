package channel

// Client -> Server
const (
	RecvMigrateIn uint16 = 20
)

// Server -> Client
const (
	SendSetField uint16 = 141
)

var RecvOpcodeNames = map[uint16]string{
	RecvMigrateIn: "MigrateIn",
}

var SendOpcodeNames = map[uint16]string{
	SendSetField: "SetField",
}
