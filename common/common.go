package common

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"net"
)

const TermID = 0x01
const ShellID = 0x02

func readUint16PrefixedData(r io.Reader) ([]byte, error) {
	var l uint16
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(io.LimitReader(r, int64(l)))
}

func ReadConnData(conn net.Conn) (int, []byte, error) {
	var connDataType uint16
	err := binary.Read(conn, binary.LittleEndian, &connDataType)
	if err != nil {
		return 0, nil, err
	}

	data, err := readUint16PrefixedData(conn)
	if err != nil {
		return 0, nil, err
	}

	return int(connDataType), data, nil
}

func WriteConnData(conn net.Conn, n int, data []byte) error {
	err := binary.Write(conn, binary.LittleEndian, uint16(n))
	if err != nil {
		return err
	}
	return writeUint16PrefixedData(conn, data)
}

func writeUint16PrefixedData(w io.Writer, data []byte) error {
	length := uint16(len(data))
	err := binary.Write(w, binary.LittleEndian, length)
	if err != nil {
		return err
	}
	_, err = w.Write(data)

	return err
}
