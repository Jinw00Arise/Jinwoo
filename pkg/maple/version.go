package maple

const (
	GameVersion uint16 = 95

	// SendVersion is XOR'd into outbound packet headers. Client expects 0xFFFF - GameVersion.
	SendVersion uint16 = 0xFFFF - GameVersion

	IVLength = 4
)
