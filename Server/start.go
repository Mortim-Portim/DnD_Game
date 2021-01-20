package Server

import (
	"flag"
	"fmt"
	"time"
	"net"
	"log"
	"os"
	"os/signal"
	"github.com/mortim-portim/GameConn/GC"
	"github.com/mortim-portim/GraphEng/GE"
	cmp "github.com/mortim-portim/GraphEng/Compression"
	"github.com/mortim-portim/TN_Engine/TNE"

	ws "github.com/gorilla/websocket"
)

const FPS = 30
const delay = time.Second/FPS

func Start() {
	flag.Parse()
	done := make(chan bool)
	wrld, err := GE.LoadWorldStructure(0, 0, 1920, 1080, *world_file, F_TILES, F_STRUCTURES)
	wrld.SetLightStats(30, 255, 2)
	wrld.SetDisplayWH(32,18)
	CheckErr(err)
	wrld_bytes, err = wrld.ToBytes()
	CheckErr(err)
	
	sm,err := TNE.GetSmallWorld(0, 0, 1920, 1080, F_TILES, F_STRUCTURES, F_ENTITY)
	CheckErr(err)
	sm.SetWorldStruct(wrld)
	SmallWorld = sm
	SmPerCon = make(map[*ws.Conn]*TNE.SmallWorld)
	
	World = TNE.GetWorld(&TNE.WorldParams{2,SmallWorld.Ef,SmallWorld.FrameCounter,wrld}, "./test")
	InitializeEntities(World)

	GC.InitSyncVarStandardTypes()
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
	
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		log.Fatal("User Termination")
		return
	}()
	time.Sleep(time.Second)
	
	for true {
		st := time.Now()
		
		*SmallWorld.FrameCounter ++
		SmallWorld.Struct.UpdateLightLevel(1)
		SmallWorld.Struct.UpdateAllLightsIfNecassary()
		World.UpdateAllPlayer()
		
		for _,sm := range(SmPerCon) {
			ok, pl := sm.HasNewActivePlayer()
			if ok {
				World.AddPlayer(pl)
				fmt.Printf("New active Player: %p\n", pl)
				PlayersChanged = true
			}
			if sm.ActivePlayer.HasPlayer() {
				x,y := sm.ActivePlayer.Player.IntPos()
				fmt.Printf("%p: %v, %v\n", sm.ActivePlayer.Player, x,y)
				cidxs := World.GetPlayerChunks(sm.ActivePlayer.Player)
				World.UpdateChunks(cidxs)
				sm.SetEntitiesFromChunks(World.Chunks, cidxs...)
			}
			sm.UpdateVars()
		}
		
		if PlayersChanged {
			for _,sm := range(SmPerCon) {
				sm.GetSyncPlayersFromWorld(World)
			}
			PlayersChanged = false
		}
		out := World.PrintPlayerPos()
		if len(out) > 0 {
			fmt.Println("waiting: ", out)
		}
		//playerJoining.Lock()
		ServerManager.UpdateSyncVars()
		Server.WaitForAllConfirmations()
		//playerJoining.Unlock()
		
		t := time.Now().Sub(st)
		if t < delay {
			time.Sleep(delay-t)
		}
	}
	
	<-done
}

func ServerInput(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Printf("Client %s send msg of len(%v): '%v'\n", c.RemoteAddr().String(), len(msg), msg)
}
func ServerNewConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("New Client Connected: ", c.RemoteAddr().String())
	
	playerJoining.Lock()
	
	data := append([]byte{GC.BINARYMSG}, []byte(TNE.NumberOfSVACIDs_Msg)...)
	data = append(data, cmp.Int16ToBytes(int16(TNE.GetSVACID_Count()))...)
	s.Send(data, c)
	s.WaitForConfirmation(c)
	newSM := SmallWorld.New()
	newSM.Register(ServerManager, c)
	newSM.SetWorldStruct(newSM.Struct)
	SmPerCon[c] = newSM
	
	playerJoining.Unlock()
}
func ServerCloseConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("Client Disconnected: ", c.RemoteAddr().String())
	if sm, ok := SmPerCon[c]; ok && sm.ActivePlayer.HasPlayer() {
		fmt.Printf("Removing Player %p from the world\n", SmPerCon[c].ActivePlayer.Player)
		World.RemovePlayer(SmPerCon[c].ActivePlayer.Player)
	}
	delete(SmPerCon, c)
	PlayersChanged = true
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
