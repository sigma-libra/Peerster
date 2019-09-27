package client

//./client -UIPort=10000 -msg=Hello

import (
	"flag"
)

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()
}