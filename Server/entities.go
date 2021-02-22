package Server

import (
	"github.com/mortim-portim/TN_Engine/TNE"
)

type EU_Random_Moves struct {
	UpdatePeriodMin, UpdatePeriodMax  int
	MovementMin, MovementMax          float64
	SpeedMin, SpeedMax                float64
	framesSinceLastUpdate, nextUpdate int
}

func (u *EU_Random_Moves) Copy() TNE.EntityUpdater {
	return &EU_Random_Moves{u.UpdatePeriodMin, u.UpdatePeriodMax, u.MovementMin, u.MovementMax, u.SpeedMin, u.SpeedMax, u.framesSinceLastUpdate, u.nextUpdate}
}
func (u *EU_Random_Moves) Reset() {
	u.nextUpdate = TNE.RandomInt(u.UpdatePeriodMin, u.UpdatePeriodMax)
}
func (u *EU_Random_Moves) Update(e *TNE.Entity, world *TNE.World) {
	if u.framesSinceLastUpdate == u.nextUpdate {
		l := TNE.RandomFloat(u.MovementMin, u.MovementMax)
		dir := TNE.GetNewRandomDirection()
		speed := TNE.RandomFloat(u.SpeedMin, u.SpeedMax)
		e.ChangeOrientation(dir)
		e.Speed = speed
		e.Move(l)
		u.Reset()
		u.framesSinceLastUpdate = -1
	}
	u.framesSinceLastUpdate++
}
func SpawnEntity(name string, w *TNE.World, x, y float64, ufnc TNE.EntityUpdater) {
	ent, err := w.Ef.GetByName(name)
	CheckErr(err)
	ent.SetMiddleTo(x, y)
	if ufnc != nil {
		ent.RegisterUpdateCallback(ufnc)
	}
	w.AddEntity(ent)
}
func InitializeEntities(w *TNE.World) {
	u_r := &EU_Random_Moves{UpdatePeriodMin: 40, UpdatePeriodMax: 100, MovementMin: 1, MovementMax: 4, SpeedMin: 1, SpeedMax: 5}
	u_r.UpdatePeriodMin = int((u_r.MovementMax / u_r.SpeedMin) * FPS)
	u_r.UpdatePeriodMax = u_r.UpdatePeriodMin * 2
	u_r.Reset()

	xstart := 9
	ystart := 9
	crNames := w.Ef.EntityNames()
	for x, name := range crNames {
		for y := 0; y < 1; y++ {
			SpawnEntity(name, w, float64(x+xstart), float64(y+ystart), u_r.Copy())
		}
	}
}
