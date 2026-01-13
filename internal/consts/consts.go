package consts

import "time"

const (
	GameVersion       uint16 = 95
	SendVersion              = 0xFFFF - GameVersion
	FieldTickInterval        = 100 * time.Millisecond
)
