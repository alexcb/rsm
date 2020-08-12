package main

import (
	"context"
	"encoding/json"
	"fmt"
	//"io"
	"net"
	"os"

	"github.com/creack/pty"
	//"github.com/hashicorp/yamux"
	"github.com/alexcb/rsm/common"
)

func handlePtyData(data []byte) {
	_, err := os.Stdout.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal data: %v\n", err)
		return
	}
}

func handleWinChangeData(ptmx *os.File, data []byte) {
	var size pty.Winsize
	err := json.Unmarshal(data, &size)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal data: %v\n", err)
		return
	}

	err = pty.Setsize(ptmx, &size)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set window size: %v\n", err)
		return
	}
}

func connectTerm(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("error closing: %v\n", err)
		}
	}()

	_, err = conn.Write([]byte{common.TermID})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			fmt.Printf("reading...\n")
			connDataType, data, err := common.ReadConnData(conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read: %v\n", err)
				break
			}
			fmt.Printf("got data %d\n", connDataType)
			switch connDataType {
			case 1:
				handlePtyData(data)
			default:
				fmt.Fprintf(os.Stderr, "unknown connDataType: %v\n", connDataType)
				break
			}
		}
		cancel()
	}()
	go func() {
		for {
			buf := make([]byte, 100)
			n, err := os.Stdin.Read(buf)
			if err != nil {
				panic(err)
			}
			buf = buf[:n]
			common.WriteConnData(conn, 1, buf)
		}
		cancel()
	}()

	//go func() {
	//	for {
	//		data, err := ReadUint16PrefixedData(stream2)
	//		if err == io.EOF {
	//			return
	//		} else if err != nil {
	//			fmt.Fprintf(os.Stderr, "failed to read data: %v\n", err)
	//			break
	//		}

	//		var size pty.Winsize
	//		err = json.Unmarshal(data, &size)
	//		if err != nil {
	//			fmt.Fprintf(os.Stderr, "failed to unmarshal data: %v\n", err)
	//			break
	//		}

	//		err = pty.Setsize(ptmx, &size)
	//		if err != nil {
	//			fmt.Fprintf(os.Stderr, "failed to set window size: %v\n", err)
	//			break
	//		}

	//	}
	//	cancel()
	//}()

	<-ctx.Done()

	fmt.Fprintf(os.Stderr, "exiting interactive debugger shell\n")
	return nil
}

func main() {
	remoteConsoleAddr := "127.0.0.1:5000"
	err := connectTerm(remoteConsoleAddr)
	if err != nil {
		panic(err)
	}
}
