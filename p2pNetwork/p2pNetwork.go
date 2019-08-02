package p2pNetwork

import (
  "fmt"
  "net"
  "strconv"
  "encoding/binary"
  "crypto/rand"
	"crypto/rsa"
  //"crypto/sha512"
  //"crypto/md5"
)

// Variables

const version uint32 = 1

const RSAbits int = 2048

type ThisPeer struct {
  roomName string
  roomPassword string
  port int
  RSAKey rsa.PrivateKey
}

type Peer struct {
  RSApublic string
  conn net.Conn
}

// Simple often used functions

func packageAddLength(pck []byte) []byte {
  ret := []byte{byte(len(pck))}
  return append(ret, []byte(pck)...)
}

// functions related to the network

func InitPeer(roomName, roomPassword string) ThisPeer {
  privKey, err := rsa.GenerateKey(rand.Reader, RSAbits)
	if err != nil {
		fmt.Println(err)
  }
  return ThisPeer{roomName, roomPassword, 0, *privKey}
}

func (p ThisPeer) handleStableConnection(conn net.Conn) {
  fmt.Println("Connection stable")
}

func (p ThisPeer) handleConnection(conn net.Conn) {
  defer conn.Close()

  //conn.SetKeepAlive(true)

  // first package
  package1 := make([]byte, 4)
  binary.LittleEndian.PutUint32(package1, version)

  roomNameBytes := []byte(p.roomName)
  //package1 = append(package1, byte(len(roomNameBytes)))
  //package1 = append(package1, roomNameBytes...)
  package1 = append(package1, packageAddLength(roomNameBytes)...)

  conn.Write(package1)

  res1 := make([]byte, 1)
  conn.Read(res1)

  password := ""
  if uint8(res1[0]) > 0 {
    res1_2 := make([]byte, uint8(res1[0]))
    conn.Read(res1_2)
    password = string(res1_2)
  }

  if (password == p.roomPassword) {
    conn.Write([]byte{0x80})
    fmt.Println("A new peer successfully connected")
  } else {
    conn.Write([]byte{0x00})
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

func (p ThisPeer) Connect(peerIP string) {
  conn, err := net.Dial("tcp", peerIP)
  if err != nil {
    fmt.Println(err)
  }

  rPackage1 := make([]byte, 5)
  conn.Read(rPackage1)

  rVersion := binary.LittleEndian.Uint32(rPackage1[:4])

  if(rVersion != version) {
    fmt.Println("failed to connect to: " + peerIP + ", due to wrong version")
    conn.Close()
  } else {
    rRoomNameLength := int(rPackage1[4])
    rPackage1_2 := make([]byte, rRoomNameLength)
    conn.Read(rPackage1_2)

    roomName := string(rPackage1_2)

    if p.roomName == "" {
      p.roomName = roomName
    }
    if p.roomName != roomName {
      fmt.Println("The connected peer (" + peerIP + ") is member of the chat room: " + roomName + ", which does not correspond with the spesified name (" + p.roomName + ")")
      fmt.Println("Stopping the connection")
      conn.Close()
    } else {
      //package2 := []byte{byte(len(p.roomPassword))}
      //package2 = append(package2, []byte(p.roomPassword)...)
      package2 := packageAddLength([]byte(p.roomPassword))

      conn.Write(package2)

      rPackage3 := make([]byte, 1)
      conn.Read(rPackage3)

      if rPackage3[0] == 0x80 {
        fmt.Println("Successfully connected to " + p.roomName + " at " + peerIP)
        p.handleStableConnection(conn)
      } else if rPackage3[0] == 0x00 {
        fmt.Println("Failed to connect to " + p.roomName + " at " + peerIP)
        fmt.Println("This error can be caused by wrong password")
        conn.Close()
      }
    }
  }
}
