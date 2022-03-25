package main

import "z4/server"

func main() {
	// TODO: Support configurable port.
	err := server.Start(6355, nil)
	if err != nil {
		panic(err)
	}
}
