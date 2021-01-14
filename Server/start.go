package Server

import (
	"flag"
	"fmt"
	"time"
	"net"

	//"strings"
	"github.com/mortim-portim/GameConn/GC"
	"github.com/mortim-portim/GraphEng/GE"
	cmp "github.com/mortim-portim/GraphEng/Compression"
	"github.com/mortim-portim/TN_Engine/TNE"

	ws "github.com/gorilla/websocket"
)

func Start() {
	flag.Parse()
	done := make(chan bool)
	wrld, err := GE.LoadWorldStructure(0, 0, 1920, 1080, *world_file, F_TILES, F_STRUCTURES)
	CheckErr(err)
	wrld_bytes, err = wrld.ToBytes()
	CheckErr(err)

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
	time.Sleep(time.Second)
	
	sm,err := TNE.GetSmallWorld(0, 0, 1920, 1080, F_TILES, F_STRUCTURES, F_ENTITY)
	CheckErr(err)
	sm.SetWorldStruct(wrld)
	SmallWorld = sm
	SmPerCon = make(map[*ws.Conn]*TNE.SmallWorld)
	
	<-done
}

func ServerInput(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Printf("Client %s send msg of len(%v): '%v'\n", c.RemoteAddr().String(), len(msg), msg)

//	if msg[0] == MAP_REQUEST {
//		fmt.Print("Sending map...")
//		s.Send(wrld_bytes, s.ConnToIdx[c])
//		s.WaitForConfirmation(s.ConnToIdx[c])
//		fmt.Print("done\n")
//	}
//
//	if msg[0] == CHAR_SEND {
//		fmt.Printf("%v %v %v %v \n", msg[1], msg[2], msg[3], msg[4])
//	}
}
func ServerNewConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("New Client Connected: ", c.RemoteAddr().String())
	
	
	data := append([]byte{GC.BINARYMSG}, []byte(TNE.NumberOfSVACIDs_Msg)...)
	data = append(data, cmp.Int16ToBytes(int16(TNE.GetSVACID_Count()))...)
	s.Send(data, s.ConnToIdx[c])
	s.WaitForConfirmation(s.ConnToIdx[c])
	
	SmPerCon[c] = SmallWorld.New()
	SmPerCon[c].Register(ServerManager, c)
	
	time.Sleep(time.Second)
	SmPerCon[c].SetWorldStruct(SmPerCon[c].Struct)
	ServerManager.UpdateSyncVars()
}
func ServerCloseConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("Client Disconnected: ", c.RemoteAddr().String())
	SmallWorld.Register(ServerManager, c)
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
