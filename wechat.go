package main

import (
	"axiom"
)

func main() {
	b := axiom.New("axiom")

	b.AddAdapter(NewWeChat(b))

	b.Register(&WeChatListener{})

	b.Start()
}



