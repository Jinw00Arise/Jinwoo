package stage

// NPC represents a spawned NPC on the stage
type NPC struct {
	ObjectID   uint32
	TemplateID int    // NPC template ID from WZ data
	X          int16
	Y          int16
	F          bool   // Facing direction (true = left)
	FH         uint16 // Foothold
	RX0        int16  // Movement range min
	RX1        int16  // Movement range max
}

// NewNPC creates a new NPC with the given parameters
func NewNPC(objectID uint32, templateID int, x, y int16, f bool, fh uint16, rx0, rx1 int16) *NPC {
	return &NPC{
		ObjectID:   objectID,
		TemplateID: templateID,
		X:          x,
		Y:          y,
		F:          f,
		FH:         fh,
		RX0:        rx0,
		RX1:        rx1,
	}
}

