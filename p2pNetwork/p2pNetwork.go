package p2pNetwork

import (
  "fmt"
  "errors"
  "net"
  "strconv"
  "encoding/binary"
  "encoding/hex"
  "encoding/json"
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
  RSApublic rsa.PublicKey
  conn net.Conn
}

type mPeer struct { // the struct to be used while sending messages
  RSApublicBytes []byte
}

// Simple often used functions

//https://sosedoff.com/2014/12/15/generate-random-hex-string-in-go.html
func RandomHex(n int) (string, error) {
  bytes := make([]byte, n)
  if _, err := rand.Read(bytes); err != nil {
    return "", err
  }
  return hex.EncodeToString(bytes), nil
}

func packageAddLength(pck []byte) []byte {
  ret := []byte{byte(len(pck))}
  return append(ret, []byte(pck)...)
}

func sendBytes(conn net.Conn, b []byte) {
  conn.Write(packageAddLength(b))
}

func sendString(conn net.Conn, s string) {
  sendBytes(conn, []byte(s))
}

func mPeerToJSON(mp mPeer) ([]byte, error) {
  b, err := json.Marshal(mp)
  if err != nil {
    return nil, err
  } else {
    return b, nil
  }
}

func (p ThisPeer) tomPeer() mPeer {
  publickey := p.RSAKey.PublicKey.N
  return mPeer{publickey.Bytes()}
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

// Incoming connection

func inCsendVersion(conn net.Conn) {
  pckg := make([]byte, 4)
  binary.LittleEndian.PutUint32(pckg, version)
  conn.Write(pckg)
}

func inCsendRoomName(conn net.Conn, roomName string) {
  sendString(conn, roomName)
}

func inCgetRoomPassword(conn net.Conn) string {
  res := make([]byte, 1)
  conn.Read(res)

  password := ""
  if uint8(res[0]) > 0 {
    res2 := make([]byte, uint8(res[0]))
    conn.Read(res2)
    password = string(res2)
  }
  return password
}

func (p ThisPeer) handleConnection(conn net.Conn) {

  inCsendVersion(conn)
  inCsendRoomName(conn, p.roomName)
  password := inCgetRoomPassword(conn)

  if (password == p.roomPassword) {
    conn.Write([]byte{0x80})
    fmt.Println("A new peer successfully connected")
    p.handleStableConnection(conn)
  } else {
    conn.Write([]byte{0x00})
    conn.Close()
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

// Outgoing connection

func outCgetVersion(conn net.Conn) uint32 {
  pckg := make([]byte, 4)
  conn.Read(pckg)

  rVersion := binary.LittleEndian.Uint32(pckg[:4])

  return rVersion
}

func outCgetRoomName(conn net.Conn) string {
  packageLength := make([]byte, 1)
  conn.Read(packageLength)
  roomName := make([]byte, int(packageLength[0]))
  conn.Read(roomName)

  return string(roomName)
}

func outCCheckPassword(conn net.Conn, password string) error {
  sendString(conn, password)

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

  rVersion := outCgetVersion(conn)

  if(rVersion != version) {
    fmt.Println("failed to connect to: " + peerIP + ", due to wrong version")
    conn.Close()
  } else {

    roomName := outCgetRoomName(conn)

    if p.roomName == "" {
      p.roomName = roomName
    }
    if p.roomName != roomName {
      fmt.Println("The connected peer (" + peerIP + ") is member of the chat room: " + roomName + ", which does not correspond with the spesified name (" + p.roomName + ")")
      fmt.Println("Stopping the connection")
      conn.Close()
    } else {
      err := outCCheckPassword(conn, p.roomPassword)

      if err == nil {
        fmt.Println("Successfully connected to " + p.roomName + " at " + peerIP)
        fmt.Println(mPeerToJSON(p.tomPeer()))
        // TODO
        // send peerdata (including username)
        // receive peerdata
        // receive other peers
        p.handleStableConnection(conn)
      } else {
        fmt.Println("Failed to connect to " + p.roomName + " at " + peerIP)
        fmt.Println("This error can be caused by wrong password")
        conn.Close()
      }
    }
  }
}
