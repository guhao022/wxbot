package main

import (
	"axiom"
)

func main() {
	b := axiom.New()

	b.AddAdapter(NewWeChat(b))

	b.Register(&WeChatListener{})

	b.Start()
}
