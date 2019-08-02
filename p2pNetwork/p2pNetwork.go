package p2pNetwork

import (
  "fmt"
  "errors"
  "net"
  "strconv"
  "encoding/binary"
  "encoding/hex"
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

//https://sosedoff.com/2014/12/15/generate-random-hex-string-in-go.html
func RandomHex(n int) (string, error) {
  bytes := make([]byte, n)
  if _, err := rand.Read(bytes); err != nil {
    return "", err
  }
  return hex.EncodeToString(bytes), nil
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
  defer conn.Close()
  fmt.Println("Connection stable")
}

func (p ThisPeer) handleConnection(conn net.Conn) {
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
    p.handleStableConnection(conn)
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

func connectgetVersion(conn net.Conn) uint32 {
  pckg := make([]byte, 4)
  conn.Read(pckg)

  rVersion := binary.LittleEndian.Uint32(pckg[:4])

  return rVersion
}

func connectgetRoomName(conn net.Conn) string {
  packageLength := make([]byte, 1)
  conn.Read(packageLength)
  roomName := make([]byte, int(packageLength[0]))
  conn.Read(roomName)

  return string(roomName)
}

func connectCheckPassword(conn net.Conn, password string) error {
  pckg := packageAddLength([]byte(password))
  conn.Write(pckg)

  res := make([]byte, 1)
  conn.Read(res)

  if (res[0] == 0x80) {
    return nil
  } else {
    return errors.New("The password was rejected")
  }
}

func (p ThisPeer) Connect(peerIP string) {
  conn, err := net.Dial("tcp", peerIP)
  if err != nil {
    fmt.Println(err)
  }

  rVersion := connectgetVersion(conn)

  if(rVersion != version) {
    fmt.Println("failed to connect to: " + peerIP + ", due to wrong version")
    conn.Close()
  } else {

    roomName := connectgetRoomName(conn)

    if p.roomName == "" {
      p.roomName = roomName
    }
    if p.roomName != roomName {
      fmt.Println("The connected peer (" + peerIP + ") is member of the chat room: " + roomName + ", which does not correspond with the spesified name (" + p.roomName + ")")
      fmt.Println("Stopping the connection")
      conn.Close()
    } else {
      err := connectCheckPassword(conn, p.roomPassword)

      if err == nil {
        fmt.Println("Successfully connected to " + p.roomName + " at " + peerIP)
        p.handleStableConnection(conn)
      } else {
        fmt.Println("Failed to connect to " + p.roomName + " at " + peerIP)
        fmt.Println("This error can be caused by wrong password")
        conn.Close()
      }
    }
  }
}
