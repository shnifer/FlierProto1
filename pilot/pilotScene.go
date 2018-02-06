package main

import (
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"log"
)

type PilotScene struct {
	*Scene
	Ship *ShipGameObject
}

func NewPilotScene(r *sdl.Renderer, ch *controlHandler) *PilotScene {
	return &PilotScene{
		Scene: NewScene(r, ch),
	}
}

func (PilotScene *PilotScene) Init() {
	BackGround := newStaticImage("background.jpg")
	PilotScene.AddObject(SceneObject(BackGround))


	Particles := newParticleSystem(DEFVAL.MainEngineMaxParticles)
	PilotScene.AddObject(SceneObject(Particles))

	//DATA INIT
	for _, starData := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: starData}
		PilotScene.AddObject(SceneObject(StarGO))
	}
	log.Println("Stars on scene",len(MNT.GalaxyData))

	Ship := newShip(Particles)
	PilotScene.Ship = Ship
	PilotScene.AddObject(SceneObject(Ship))

	startLoc := PilotScene.GetObjByID(DEFVAL.StartLocationName)
	if startLoc!=nil{
		pos,_:=startLoc.(HugeMass).GetGravState()
		Ship.pos = pos.Add( DEFVAL.StartLocationOffset)
	} else {
		Ship.pos = DEFVAL.StartLocationOffset
	}

	FrontCabin := newStaticImage("cabinBorder.png")
	PilotScene.AddObject(SceneObject(FrontCabin))

	PilotScene.Scene.Init()
}

//Возвращает силу тяжести, точнее ускорение для заданной массы и заданного пробного положения
func GravityForce (attractor HugeMass, body V2.V2) V2.V2 {
	pos, mass := attractor.GetGravState()
	ort := V2.Sub(pos, body).Normed()
	dist2 := V2.Sub(pos, body).LenSqr() + DEFVAL.GravityDepthSqr
	Amp := DEFVAL.GravityConst * mass / dist2
	force := ort.Mul(Amp)
	return force
}

func (ps *PilotScene) Update(dt float32) {
	//ФИЗИКА
	s := ps.Scene
	for _, obj := range s.Objects {
		attractor, ok := obj.(HugeMass)
		if !ok {
			continue
		}
		force := GravityForce(attractor,ps.Ship.pos)
		ps.Ship.ApplyForce(force)
	}

	s.Update(dt)

	//Сдвинули камеру
	ps.CameraCenter = ps.Ship.pos

}

func (ps PilotScene) Draw() {

	s := ps.Scene
	s.Draw()

	GizmoGravityForceK:=DEFVAL.GizmoGravityForceK
	//Отрисовка "Гизмосов" гравитации
	if DEFVAL.ShowGizmoGravityRound || DEFVAL.ShowGizmoGravityForce {
		sumForce := V2.V2{}

		for _, obj := range s.Objects {
			attractor, ok := obj.(HugeMass)
			if !ok {
				continue
			}

			pos,mass:=attractor.GetGravState()

			if DEFVAL.ShowGizmoGravityForce {
				// Гизмос Наш вектор
				force := GravityForce(attractor, ps.Ship.pos).Mul(GizmoGravityForceK)

				s.R.SetDrawColor(0, 0, 255, 255)
				s.R.DrawLine(winW/2, winH/2, winW/2+int32(force.X), winH/2+int32(force.Y))
				sumForce = sumForce.Add(force)
			}

			if DEFVAL.ShowGizmoGravityRound {
				dotsInCirle := DEFVAL.GizmoGravityRoundDotsInCirle
				var GizmoGravLevels= DEFVAL.GizmoGravityRoundLevels
				//Гизмос вокруг планеты

				levelsCount := len(GizmoGravLevels)

				points := make([]sdl.Point, dotsInCirle+1)

				for level := 0; level < levelsCount; level++ {

					//GizmoLevel - сила(ускорение)
					//GizmoLevel = GravityConst*mass/RadSqr
					GravRadSqr := DEFVAL.GravityConst*mass / GizmoGravLevels[level]
					GravRad := float32(math.Sqrt(float64(GravRadSqr)))
					rect := f32Rect{pos.X - GravRad, pos.Y - GravRad, 2 * GravRad, 2 * GravRad}
					_, inCamera := ps.CameraTransformRect(rect)
					if !inCamera {
						continue
					}

					//n+1 чтобы замкнуть круг
					for a := 0; a <= dotsInCirle; a++ {
						dot := pos.AddMul(V2.InDir(float32(a*360/dotsInCirle)), GravRad)
						x, y := ps.CameraTransformV2(dot)
						points[a] = sdl.Point{x, y}
					}
					s.R.SetDrawColor(128, 128, 128, 128)
					s.R.DrawLines(points)
				}
			}
		}

		if DEFVAL.ShowGizmoGravityForce {
			//Гизмос наш суммарный вектор
			s.R.SetDrawColor(0, 255, 0, 255)
			s.R.DrawLine(winW/2, winH/2, winW/2+int32(sumForce.X), winH/2+int32(sumForce.Y))
		}
	}
}