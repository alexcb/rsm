package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/alexcb/rsm/common"
	"github.com/creack/pty"
)

var (
	// Version is the version of the debugger
	Version string

	// ErrNoShellFound occurs when the container has no shell
	ErrNoShellFound = fmt.Errorf("no shell found")
)

func getShellPath() (string, bool) {
	for _, sh := range []string{
		"bash", "ksh", "zsh", "sh",
	} {
		if path, err := exec.LookPath(sh); err == nil {
			return path, true
		}
	}
	return "", false
}

func handlePtyData(ptmx *os.File, data []byte) {
	_, err := ptmx.Write(data)
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

func interactiveMode(remoteConsoleAddr string) error {
	conn, err := net.Dial("tcp", remoteConsoleAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("error closing: %v\n", err)
		}
	}()

	_, err = conn.Write([]byte{common.ShellID})
	if err != nil {
		return err
	}

	shellPath, ok := getShellPath()
	if !ok {
		return ErrNoShellFound
	}
	c := exec.Command(shellPath)

	ptmx, e := pty.Start(c)
	if e != nil {
		fmt.Printf("failed to start pty: %v\n", e)
		return e
	}
	defer func() { _ = ptmx.Close() }() // Best effort.

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			connDataType, data, err := common.ReadConnData(conn)
			if err != nil {
				break
			}
			fmt.Printf("got data\n")
			switch connDataType {
			case 1:
				handlePtyData(ptmx, data)
			case 2:
				handleWinChangeData(ptmx, data)
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
			n, err := ptmx.Read(buf)
			if err != nil {
				panic(err)
			}
			buf = buf[:n]
			fmt.Printf("sending data\n")
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
	err := interactiveMode(remoteConsoleAddr)
	if err != nil {
		panic(err)
	}
}
