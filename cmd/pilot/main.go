package main

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/Shnifer/flierproto1/fps"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"runtime"
	"time"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"fmt"
)

//Константы экрана
//TODO:вынести параметры экрана во внешний файл конфигурации
var winW int32
var winH int32

const ResourcePath = "res/"

type GameState byte

const (
	//TODO: Экран перезагрузки
	state_Login GameState = iota
	state_PilotSpace
	state_NaviSpace
)

//Пока время синхронизации храним для звёзд в глобальной переменной
//PILOT отсчитывает и передаёт, NAVI и CARGO принимают
var GlobalNetSessionTime float32


func main() {

	runtime.LockOSThread()

	deferMe, renderer, Joystick := InitSomeShit()
	defer deferMe()

	ControlHandler := control.NewControlHandler(Joystick)

	PilotScene := NewPilotScene(renderer, ControlHandler)
	PilotScene.Init()

	//Проверочный показывать фпс, он же заглушка на систему каналов
	initFPS := fps.InitStruct{
		MIN_FRAME_MS:           DEFVAL.MIN_FRAME_MS,
		MIN_PHYS_MS:            DEFVAL.MIN_PHYS_MS,
		MAX_FRAME_MS:           DEFVAL.MAX_FRAME_MS,
		MAX_PHYS_MS:            DEFVAL.MAX_PHYS_MS,
		FPS_UPDATE_MS:          DEFVAL.FPS_UPDATE_MS,
		TickerBalancerOverhead: DEFVAL.TickerBalancerOverhead,
	}
	ShowFpsTick, fpsControl := fps.Start(initFPS)
	defer close(fpsControl)

	lastPhysFrame := time.Now()

	graphFrameN, physFrameN, ioFrameN, netFrameN := 0, 0, 0, 0
	var maxDt, maxGraphT, maxPhysT float32

	breakMainLoop := make(chan bool, 1)

	IOTick := time.Tick(15 * time.Millisecond)
	NetTick := time.Tick(50*time.Millisecond)

	GlobalNetSessionTime = 0

loop:
	for {
		select {
		//команда на выход
		case <-breakMainLoop:
			break loop
		//Время передать фпс
		case <-ShowFpsTick:
			fpsControl <- fps.FpsData{graphFrameN, physFrameN, ioFrameN, netFrameN,
				maxDt, maxGraphT, maxPhysT}
			maxDt = 0.0
			maxGraphT = 0.0
			maxPhysT = 0.0

		//ПРИОРИТЕТ 1: тик ФИЗИЧЕСКОГО движка
		case <-fps.PTick:
			deltaTime := float32(time.Since(lastPhysFrame).Seconds())
			if deltaTime > maxDt {
				maxDt = deltaTime
			}
			lastPhysFrame = time.Now()
			physFrameN++

			GlobalNetSessionTime +=deltaTime

			ControlHandler.BeforeUpdate()
			PilotScene.Update(deltaTime)
			T := float32(time.Since(lastPhysFrame).Seconds())
			if T > maxPhysT {
				maxPhysT = T
			}
		default:
			select {
			//ПРИОРИТЕТ 2: тик ГРАФИЧЕСКОГО движка
			case <-fps.GTick:
				graphFrameN++
				start := time.Now()
				renderer.Clear()
				PilotScene.Draw()
				renderer.Present()
				T := float32(time.Since(start).Seconds())
				if T > maxGraphT {
					maxGraphT = T
				}

			//ПРИОРИТЕТ 3: снятие состояния УПРАВЛЕНИЯ
			case <-IOTick:
				ioFrameN++
				DoMainLoopIO(breakMainLoop, ControlHandler)
			//ПРИОРИТЕТ 3: обновление данных с СЕТЬЮ
			case <-NetTick:
				netFrameN++
				DoMainLoopNet(PilotScene)
			default:
			}
		}
	}
}

func DoMainLoopIO(breakMainLoop chan bool, handler *control.Handler) {
	//Проверяем и хэндлим события СДЛ. Выход -- обязательно, а то не закроется
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch ev := event.(type) {
		case *sdl.QuitEvent:
			log.Println("quit")
			breakMainLoop <- true
		case *sdl.KeyboardEvent:
			//Кнопку выходи обрабатываем здесь, чтобы порвать главный цикл
			scan := ev.Keysym.Scancode
			if scan == sdl.SCANCODE_ESCAPE {
				breakMainLoop <- true
			}
		default:
		}
		handler.HandleSDLEvent(event)
	}
	handler.IOUpdate()
}

func DoMainLoopNet(scene *PilotScene) {
	shipData:=MNT.ShipPosData{
		Pos: scene.Ship.pos,
		Speed: scene.Ship.speed,
		Angle: scene.Ship.angle,
		AngleSpeed: scene.Ship.angleSpeed,
	}
	params:=MNT.EncodeShipPos(shipData)
	MNT.SendBroadcast(MNT.SHIP_POS, params)
	MNT.SendBroadcast(MNT.SESSION_TIME, fmt.Sprintf("%f", GlobalNetSessionTime))
}


func timeCheck(caption string) func() {
	Start := time.Now()
	return func() {
		t:=time.Now()
		log.Println(caption, t.Sub(Start).Seconds()*1000000, "micro s")
	}
}