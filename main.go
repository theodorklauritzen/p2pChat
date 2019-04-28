package main

import (
  "./p2pNetwork"
  //"fmt"
  "flag"
)

// main <PORT> <roomName> [password] [options]
//
// [-c <ip>]

func main() {

  var port int
  flag.IntVar(&port, "p", 8000, "listening port")

  var roomName string
  flag.StringVar(&roomName, "n", "", "The room name")

  var roomPass string
  flag.StringVar(&roomPass, "s", "", "The room name or room secret")

  var connectIP string
  flag.StringVar(&connectIP, "c", "", "Connect to another peer")

  flag.Parse()

  p := p2pNetwork.InitPeer(roomName, roomPass)
  go p.Listen(port)

  if connectIP != "" {
    p.Connect(connectIP)
  }

  for {}
}
