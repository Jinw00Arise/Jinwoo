// Package field provides map instance management.
package field

import "sync/atomic"

var objectIDCounter uint32 = 999 // Start at 1000 after first increment

// NextObjectID returns the next unique object ID.
func NextObjectID() uint32 {
	return atomic.AddUint32(&objectIDCounter, 1)
}

// NPC implements the game.FieldNPC interface.
type NPC struct {
	objectID   uint32
	templateID int
	x, y       int16
	facesRight bool
}

// NewNPC creates a new field NPC.
func NewNPC(templateID int, x, y int16, facesRight bool) *NPC {
	return &NPC{
		objectID:   NextObjectID(),
		templateID: templateID,
		x:          x,
		y:          y,
		facesRight: facesRight,
	}
}

func (n *NPC) ObjectID() uint32    { return n.objectID }
func (n *NPC) TemplateID() int     { return n.templateID }
func (n *NPC) Position() (int16, int16) { return n.x, n.y }
func (n *NPC) Facing() bool        { return n.facesRight }

// Portal implements the game.Portal interface.
type Portal struct {
	id           int
	name         string
	portalType   int
	x, y         int16
	targetMap    int
	targetPortal string
	script       string
}

// NewPortal creates a new portal.
func NewPortal(id int, name string, portalType int, x, y int16, targetMap int, targetPortal, script string) *Portal {
	return &Portal{
		id:           id,
		name:         name,
		portalType:   portalType,
		x:            x,
		y:            y,
		targetMap:    targetMap,
		targetPortal: targetPortal,
		script:       script,
	}
}

func (p *Portal) ID() int                   { return p.id }
func (p *Portal) Name() string              { return p.name }
func (p *Portal) Type() int                 { return p.portalType }
func (p *Portal) Position() (int16, int16)  { return p.x, p.y }
func (p *Portal) TargetMap() int            { return p.targetMap }
func (p *Portal) TargetPortal() string      { return p.targetPortal }
func (p *Portal) Script() string            { return p.script }

