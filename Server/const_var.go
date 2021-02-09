package Server

import (
	"sync"
	"flag"
	ws "github.com/gorilla/websocket"
	"github.com/mortim-portim/GameConn/GC"
	"github.com/mortim-portim/TN_Engine/TNE"
)

const (
	RES            = "./.res"
	F_KEYLI_MAPPER = RES + "/keyli.txt"
	F_ICONS        = RES + "/Icons"
	F_MAPS         = RES + "/Maps"
	F_CHARACTER    = RES + "/Character"
	F_STRUCTURES   = F_MAPS + "/structures"
	F_TILES        = F_MAPS + "/tiles"

	F_AUDIO      = RES + "/Audio"
	F_SOUNDTRACK = F_AUDIO + "/Soundtrack"

	F_IMAGES        = RES + "/Images"
	F_GUI           = F_IMAGES + "/GUI"
	F_TITLESCREEN   = F_GUI + "/Titlescreen"
	F_PLAYMENU      = F_GUI + "/PlayMenu"
	F_CHARACTERMENU = F_GUI + "/CharacterMenu"
	F_BUTTONS       = F_GUI + "/Buttons"
	F_CONNECTING    = F_GUI + "/Connecting"

	F_ENTITY   = RES + "/Entities"

	MAP_REQUEST = GC.MESSAGE_TYPES + 0
	CHAR_SEND   = GC.MESSAGE_TYPES + 1
)

var (
	port       = flag.String("port", "8080", "Port of the server to run on")
	world_file = flag.String("world", "./.res/Maps/Worlds/benTestMap1.map", "path of the world that the server is going to host")
	Server     *GC.Server
	ServerManager *GC.ServerManager
	SmallWorld *TNE.SmallWorld
	
	World *TNE.World
	PlayersChanged, UpdateAllPositions bool
	ActionReset *chan bool
	
	playerJoining sync.Mutex
	
	SmPerCon map[*ws.Conn]*TNE.SmallWorld
	
	wrld_bytes []byte
)
