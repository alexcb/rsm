package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/alexcb/rsm/common"
)

func TestServer(t *testing.T) {
	addr := "127.0.0.1:9834"
	s := NewServer(addr)
	go s.Start()
	defer s.Stop()

	time.Sleep(10 * time.Millisecond)

	// first open terminal
	termConn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	_, err = termConn.Write([]byte{common.TermID})
	if err != nil {
		panic(err)
	}

	// then the shell terminal
	shellConn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	_, err = shellConn.Write([]byte{common.ShellID})
	if err != nil {
		panic(err)
	}

	inputStr := "hello world"

	// send data from shell to term
	_, err = shellConn.Write([]byte(inputStr))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 100)
	n, err := termConn.Read(buf)
	outputStr := string(buf[:n])

	if inputStr != outputStr {
		t.Error(fmt.Sprintf("want %v; got %v", inputStr, outputStr))
	}

}
