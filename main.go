package main

import (
  "./p2pNetwork"
)

func main() {
  p := p2pNetwork.InitPeer("room", "pass")
  p.Listen(8080)
}
