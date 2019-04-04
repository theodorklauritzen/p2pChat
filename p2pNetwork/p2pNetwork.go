package p2pNetwork

import (
  "fmt"
  "net"
  "strconv"
)

type Peer struct {
  port int
}

func NewPeer() Peer {
  return Peer{}
}

func (p Peer) handleConnection(conn net.Conn) {
  defer conn.Close()

  conn.Write([]byte{0x65, 0x65, 0x65})
}

func (p Peer) Listen(port int) {
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
