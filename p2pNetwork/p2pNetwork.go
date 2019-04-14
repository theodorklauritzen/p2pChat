package p2pNetwork

import (
  "fmt"
  "net"
  "strconv"
  "encoding/binary"
)

const version uint = 1

type ThisPeer struct {
  roomName string
  roomPassword string
  port int
}

func InitPeer(roomName, roomPassword string) ThisPeer {
  return ThisPeer{roomName, roomPassword, 0}
}

func (p ThisPeer) handleConnection(conn net.Conn) {
  defer conn.Close()

  //conn.SetKeepAlive(true)

  // first package
  package1 := make([]byte, 4)
  binary.LittleEndian.PutUint32(package1, uint32(version))

  roomNameBytes := []byte(p.roomName)
  package1 = append(package1, byte(len(roomNameBytes)))
  package1 = append(package1, roomNameBytes...)

  conn.Write(package1)

  res1 := make([]byte, 256)
  conn.Read(res1)
  password := string(res1[1:])
  fmt.Println(password)

  if (password == p.roomPassword) {
    conn.Write([]byte{0x65, 0x65, 0x65})
  } else {
    conn.Write([]byte("Access Denied: wrong password"))
  }
}

func (p ThisPeer) Listen(port int) {
  p.port = port
  ln, err := net.Listen("tcp", ":" + strconv.Itoa(port))
  if err != nil {
  	fmt.Println(err)
  }
  for {
  	conn, err := ln.Accept()
  	if err != nil {
  		fmt.Println(err)
  	}
  	go p.handleConnection(conn)
  }
}
