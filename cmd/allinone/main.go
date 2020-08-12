package main

import (
	"fmt"
	"net"
	"time"

	"github.com/alexcb/rsm/common"
	"github.com/alexcb/rsm/server"
)

const addr = "127.0.0.1:5001"

func shell() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("error closing: %v\n", err)
		}
	}()

	// identify we are the shell
	conn.Write([]byte{common.ShellID})

	conn.Write([]byte("shell data"))

	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	buf = buf[:n]
	fmt.Printf("got back: %v\n", buf)
}

func term() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("error closing: %v\n", err)
		}
	}()

	// identify we are the term
	conn.Write([]byte{common.TermID})

	conn.Write([]byte("term data"))

	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	buf = buf[:n]
	fmt.Printf("got back: %v\n", buf)
}

func main() {
	x := server.NewServer(addr)

	go x.Start()

	time.Sleep(time.Millisecond)
	go shell()

	time.Sleep(time.Millisecond)
	go term()

	time.Sleep(1 * time.Second)
	x.Stop()
}
