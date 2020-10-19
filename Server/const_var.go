package Server

import (
	"marvin/GraphEng/GE"
	"marvin/GameConn/GC"
	"flag"
)

const (
	RES = 						"./.res"
	STRUCTURE_FILES = 			"/Maps/structures"
	TILE_FILES = 				"/Maps/tiles"
	
	MAP_REQUEST = 				GC.MESSAGE_TYPES+0
)
var (
	port = flag.String("port", "8080", "Port of the server to run on")
	world_file = flag.String("world", "./.res/Maps/Worlds/ben.map", "path of the world that the server is going to host")
	Server *GC.Server
	World *GE.WorldStructure
	wrld_bytes []byte
	done chan struct{}
)