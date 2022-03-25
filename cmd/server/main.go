package main

import "z4/server"

func main() {
	err := server.Start(6355, nil)
	if err != nil {
		panic(err)
	}
}
