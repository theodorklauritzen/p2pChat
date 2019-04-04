package main

import (
  "./p2pNetwork"
)

func main() {
  p := p2pNetwork.NewPeer()
  p.Listen(8080)
}
