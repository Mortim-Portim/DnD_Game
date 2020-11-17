package Server

import (
	"flag"
	"fmt"
	"log"
	"net"

	//"strings"
	"marvin/GameConn/GC"
	"marvin/GraphEng/GE"

	ws "github.com/gorilla/websocket"
)

func Start() {
	flag.Parse()
	if flag.NFlag() != 2 {
		log.Fatal("Please specifiy the following flags: [port], [port]\nUse -help to display all posible flags")
	}
	done = make(chan struct{})

	wrld, err := GE.LoadWorldStructure(0, 0, 1920, 1080, *world_file, RES+TILE_FILES, RES+STRUCTURE_FILES)
	CheckErr(err)
	World = wrld
	wrld_bytes, err = wrld.ToBytes()
	wrld_bytes = append([]byte{MAP_REQUEST}, wrld_bytes...)
	CheckErr(err)

	GC.InitSyncVarStandardTypes()
	Server = GC.GetNewServer()
	servermanager := GC.GetServerManager(Server)
	servermanager.InputHandler = ServerInput
	servermanager.OnNewConn = ServerNewConn
	servermanager.OnCloseConn = ServerCloseConn

	//Runs the server
	ipAddr := GetLocalIP()
	ipAddrS := fmt.Sprintf("%s:%s", ipAddr, *port)
	fmt.Println("Running on:", ipAddrS)
	Server.Run(ipAddrS)

	<-done
}

func ServerInput(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Printf("Client %s send msg of len(%v): '%v'\n", c.RemoteAddr().String(), len(msg), msg)

	if msg[0] == MAP_REQUEST {
		fmt.Print("Sending map...")
		s.Send(wrld_bytes, s.ConnToIdx[c])
		s.WaitForConfirmation(s.ConnToIdx[c])
		fmt.Print("done\n")
	}

	if msg[0] == CHAR_SEND {
		fmt.Printf("%v %v %v %v \n", msg[1], msg[2], msg[3], msg[4])
	}
}
func ServerNewConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("New Client Connected: ", c.RemoteAddr().String())
}
func ServerCloseConn(c *ws.Conn, mt int, msg []byte, err error, s *GC.Server) {
	fmt.Println("Client Disconnected: ", c.RemoteAddr().String())

}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

//// Get preferred outbound ip of this machine
//func GetOutboundIP() string {
//    conn, err := net.Dial("udp", "8.8.8.8:80")
//    if err != nil {
//        log.Fatal(err)
//    }
//    defer conn.Close()
//
//    localAddr := strings.Split(conn.LocalAddr().String(), ":")[0]
//    return localAddr
//}

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
