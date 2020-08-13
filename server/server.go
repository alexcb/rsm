package server

import (
	"context"
	"fmt"
	"net"
	"os"
	//"os/signal"
	"sync"
	//"syscall"
	//"github.com/creack/pty"
)

type Server struct {
	shellConn    net.Conn
	terminalConn net.Conn
	mux          sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	dataForShell    chan []byte
	dataForTerminal chan []byte

	sigs chan os.Signal
	addr string
}

//func (s *session) sendNewWindowSize(size *pty.Winsize) error {
//	b, err := json.Marshal(size)
//	if err != nil {
//		return err
//	}
//	return writeUint16PrefixedData(s.resizeConn, b)
//}
//
//func (s *session) handle() error {
//	for {
//		stream, err := s.yaSession.Accept()
//		if err != nil {
//			return err
//		}
//
//		buf := make([]byte, 1)
//		stream.Read(buf)
//
//		switch buf[0] {
//		case 0x01:
//			go s.handle1(stream)
//
//		case 0x02:
//			s.resizeConn = stream
//			s.server.sigs <- syscall.SIGWINCH
//		default:
//			return fmt.Errorf("unsupported stream code %v", buf[0])
//		}
//	}
//}
//
//func (s *session) handle1(conn net.Conn) error {
//	go func() {
//		_, _ = io.Copy(os.Stdout, conn)
//		s.cancel()
//	}()
//	go func() {
//		_, _ = io.Copy(conn, os.Stdin)
//		s.cancel()
//	}()
//
//	<-s.ctx.Done()
//	return nil
//}

func (s *Server) handleConn(conn net.Conn, readFrom, writeTo chan []byte) {
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("err %v\n", err)
				break
			}
			fmt.Printf("read from %v; send to chan\n", conn)
			writeTo <- buf[:n]
			fmt.Printf("send to chan done\n")
		}
		cancel()
	}()

	go func() {
		defer wg.Done()

	outer:
		for {
			select {
			case <-ctx.Done():
				break outer
			case data := <-readFrom:
				fmt.Printf("read from chan; writting to conn %v %d bytes\n", conn, len(data))
				_, err := conn.Write(data)
				if err != nil {
					fmt.Printf("failed %v\n", err)
				}
				fmt.Printf("wrote done\n")
			}
		}
	}()

	wg.Wait()
}

func (s *Server) handleTermConn(conn net.Conn) {
	s.handleConn(conn, s.dataForShell, s.dataForTerminal)
}

func (s *Server) handleShellConn(conn net.Conn) {
	s.handleConn(conn, s.dataForTerminal, s.dataForShell)
}

func (s *Server) handleRequest(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1)
	conn.Read(buf)

	var isShellConn bool
	switch buf[0] {
	case 0x01:
		isShellConn = true
	case 0x02:
		isShellConn = false
	default:
		fmt.Fprintf(os.Stderr, "unexpected data")
		return
	}

	err := func() error {
		s.mux.Lock()
		defer s.mux.Unlock()
		if isShellConn {
			if s.shellConn != nil {
				return fmt.Errorf("shell already connected")
			}
			s.shellConn = conn
		} else {
			if s.terminalConn != nil {
				return fmt.Errorf("terminal already connected")
			}
			s.terminalConn = conn
		}
		return nil
	}()

	if err != nil {
		panic(err)
		return
	}

	if isShellConn {
		s.handleShellConn(conn)
	} else {
		s.handleTermConn(conn)
	}

	//yaSession, err := yamux.Server(conn, nil)
	//if err != nil {
	//	return err
	//}

	//ctx, cancel := context.WithCancel(context.Background())
	//s.session = &session{
	//	yaSession: yaSession,
	//	ctx:       ctx,
	//	cancel:    cancel,
	//	server:    s,
	//}
	//defer cancel()
	//defer func() { s.session = nil }()

	//oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	//if err != nil {
	//	return err
	//}
	//defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }()

	//return s.session.handle()
}

//func (s *server) windowResizeHandler() error {
//	for {
//		select {
//		case _ = <-s.sigs:
//
//		case <-s.ctx.Done():
//			return nil
//		}
//		if len(s.sigs) > 0 {
//			continue
//		}
//		size, err := pty.GetsizeFull(os.Stdin)
//		if err != nil {
//			fmt.Printf("failed to get size: %v\n", err)
//		} else {
//			if s.session != nil {
//				s.session.sendNewWindowSize(size)
//			}
//		}
//	}
//}

// Start starts the debug server listener
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	go func() {
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accepting: %v", err.Error())
				continue
			}
			go s.handleRequest(conn)
		}
	}()
	return nil
}

func (s *Server) Stop() error {
	//s.cancel()
	return nil
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,

		dataForShell:    make(chan []byte, 100),
		dataForTerminal: make(chan []byte, 100),
	}
}
