package main

import (
	"time"

	"github.com/alexcb/rsm/server"
)

const addr = "127.0.0.1:5000"

func main() {
	x := server.NewServer(addr)
	x.Start()
	time.Sleep(time.Hour)
	//x.Stop()
}
