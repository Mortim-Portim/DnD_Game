package Server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/mortim-portim/GameConn/GC"
	"github.com/mortim-portim/GraphEng/GE"
	cmp "github.com/mortim-portim/GraphEng/compression"
	"github.com/mortim-portim/TN_Engine/TNE"
)

var FPS = TNE.FPS

var delay = time.Duration(float64(time.Second) / FPS)

func onUnexpectedError() {
	if r := recover(); r != nil {
		CloseServer("unexpected Error:", r, "\n", string(debug.Stack()))
	}
}
func Start() {
	playerJoining.Lock()
	flag.Parse()
	GE.Init("", FPS)
	GE.StartProfiling(cpuprofile)
	defer onUnexpectedError()
	done := make(chan bool)
	wrld, err := GE.LoadWorldStructure(0, 0, 1920, 1080, *world_file, F_TILES, F_STRUCTURES)
	if err != nil {
		fmt.Println("Error loading worldstructure generating one")
		wrld = GE.GetWorldStructure(0, 0, 1920, 1080, 100, 100, 32, 18)
		wrld.LoadTiles(F_TILES)
		wrld.LoadStructureObjs(F_STRUCTURES)
		wrld.TileMat.FillAll(1)
		wrld.SetMiddle(0, 0, true)

	}
	wrld.SetLightStats(30, 220)
	wrld.SetDisplayWH(32, 18)

	wrld_bytes = wrld.ToBytes()

	sm, err := TNE.GetSmallWorld(0, 0, 1920, 1080, F_TILES, F_STRUCTURES, F_ENTITY)
	CheckErr(err)
	sm.SetWorldStruct(wrld)
	sm.FrameChanSendPeriod = int(FPS * 20)
	sm.SetTimePerFrame(int64(float64(time.Hour) / FPS))
	SmallWorld = sm

	SmPerCon = make(map[*ws.Conn]*TNE.SmallWorld)

	World = TNE.GetWorld(&TNE.WorldParams{2, SmallWorld.Ef, SmallWorld.FrameCounter, wrld}, "./test")
	InitializeEntities(World)
	SmallWorld.World = World

	GC.InitSyncVarStandardTypes()
	GC.PRINT_LOG_PRIORITY = 3
	Server = GC.GetNewServer()
	ServerManager = GC.GetServerManager(Server)
	ServerManager.InputHandler = ServerInput
	ServerManager.OnNewConn = ServerNewConn
	ServerManager.OnCloseConn = ServerCloseConn
	//Runs the server
	ipAddr := GetLocalIP()
	ipAddrS := fmt.Sprintf("%s:%s", ipAddr, *port)
	fmt.Println("Running on:", ipAddrS)
	Server.Run(ipAddrS)

	Server.InputWaiting = true

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		CloseServer()
		return
	}()
	time.Sleep(time.Second)
	playerJoining.Unlock()

	for true {
		//fmt.Println("---------------------------------------Stop time")
		st := time.Now()

		playerJoining.Lock()

		//fmt.Println("---------------------------------------Receive Input first")
		Server.HandleInput()

		//fmt.Println("---------------------------------------Update World")
		*SmallWorld.FrameCounter++
		World.UpdateLights(time.Duration(SmallWorld.TimePerFrame))
		World.UpdateAllPlayer()

		//fmt.Println("---------------------------------------Update smallworlds with entities")
		for _, sm := range SmPerCon {
			ok, pl := sm.HasNewActivePlayer()
			if ok {
				World.AddPlayer(pl)
				sm.SetActivePlayerID(int16(EntityIDFactory.GetID(0, 32768)))
				fmt.Printf("New active Player: %p\n", pl)
				PlayersChanged = true
			}
			if sm.ActivePlayer.HasPlayer() {
				cidxs := World.GetPlayerChunks(sm.ActivePlayer.Player)
				World.UpdateChunks(cidxs)
				sm.SetEntitiesFromChunks(World.Chunks, cidxs...)
			}
		}
		//fmt.Println("---------------------------------------Update the player of the smallworlds")
		if PlayersChanged {
			World.UpdateAllPos()
			for _, sm := range SmPerCon {
				sm.GetSyncPlayersFromWorld(World)
			}
			PlayersChanged = false
		}
		//fmt.Println("---------------------------------------Set SyncVars of the smallworlds")
		for _, sm := range SmPerCon {
			sm.UpdateVars()
		}

		//fmt.Println("---------------------------------------send syncvars buffered")
		ServerManager.UpdateSyncVarsBuffered()

		msg, num := World.Print(false)
		if num > 0 {
			fmt.Println(msg)
		}
		//fmt.Println("---------------------------------------reset applied actions")
		World.ResetActions()

		playerJoining.Unlock()

		//fmt.Println("---------------------------------------Wait up to 33.33 ms (to update at 30 FPS)")
		t := time.Now().Sub(st)
		if t < delay {
			time.Sleep(delay - t)
		}
	}
	<-done
}
func CloseServer(msg ...interface{}) {
	GE.StopProfiling(cpuprofile, memprofile)
	log.Fatal("Termination: ", fmt.Sprint(msg...))
}
func ServerInput(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	//fmt.Printf("Client %s send msg of len(%v): '%v'\n", c.RemoteAddr().String(), len(msg), msg)
}
func ServerNewConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("New Client Connected: ", c.RemoteAddr().String())

	playerJoining.Lock()
	data := append([]byte{GC.BINARYMSG}, []byte(TNE.NumberOfSVACIDs_Msg)...)
	data = append(data, cmp.Int16ToBytes(int16(TNE.GetSVACID_Count()))...)
	s.SendBuffered(data, c)
	s.WaitForConfirmation(c)

	newSM := SmallWorld.New()
	newSM.Register(ServerManager, c)

	newSM.SetWorldStruct(newSM.Struct)
	newSM.SetTimePerFrame(SmallWorld.TimePerFrame)
	ServerManager.UpdateSyncVarsNormal()
	s.WaitForConfirmation(c)

	SmPerCon[c] = newSM
	PlayersChanged = true
	playerJoining.Unlock()
}
func ServerCloseConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("Client Disconnected: ", c.RemoteAddr().String())

	playerJoining.Lock()
	if sm, ok := SmPerCon[c]; ok && sm.ActivePlayer.HasPlayer() {
		fmt.Printf("Removing Player %p from the world\n", SmPerCon[c].ActivePlayer.Player)
		EntityIDFactory.AddID(int(SmPerCon[c].ActivePlayer.ID))
		World.RemovePlayer(SmPerCon[c].ActivePlayer.Player)
	}
	delete(SmPerCon, c)
	PlayersChanged = true
	playerJoining.Unlock()

}
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
