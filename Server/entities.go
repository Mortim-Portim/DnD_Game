package Server

import (
	"github.com/mortim-portim/TN_Engine/TNE"
)
type EU_Random_Moves struct {
	UpdatePeriodMin, UpdatePeriodMax int
	MovementMin, MovementMax float64
	FrmsMin, FrmsMax int
	framesSinceLastUpdate, nextUpdate int
}
func (u *EU_Random_Moves) Copy() *EU_Random_Moves {
	return &EU_Random_Moves{u.UpdatePeriodMin, u.UpdatePeriodMax, u.MovementMin, u.MovementMax, u.FrmsMin, u.FrmsMax, u.framesSinceLastUpdate, u.nextUpdate}
}
func (u *EU_Random_Moves) Reset() {
	u.nextUpdate = TNE.RandomInt(u.UpdatePeriodMin, u.UpdatePeriodMax)
}
func (u *EU_Random_Moves) Update(e *TNE.Entity, world *TNE.World) {
	if u.framesSinceLastUpdate == u.nextUpdate {
		l := TNE.RandomFloat(u.MovementMin, u.MovementMax)
		dir := TNE.GetNewRandomDirection()
		frms := TNE.RandomInt(u.FrmsMin, u.FrmsMax)
		e.ChangeOrientation(dir)
		e.Move(l, frms)
		u.Reset()
		u.framesSinceLastUpdate = -1
	}
	u.framesSinceLastUpdate ++
}

func InitializeEntities(w *TNE.World) {
	u_r := &EU_Random_Moves{UpdatePeriodMin:40, UpdatePeriodMax:100, MovementMin:1, MovementMax:4, FrmsMin:20, FrmsMax:40}
	u_r.Reset()
	
	halfling, err := w.Ef.GetByName("Halfling")
	CheckErr(err)
	halfling.SetMiddleTo(5,5)
	halfling.RegisterUpdateCallback(u_r.Copy())
	
	w.AddEntity(halfling)
}