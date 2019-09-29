package udp

// sources: https://holwech.github.io/blog/Creating-a-simple-UDP-module/
import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dedis/protobuf"
)

const checkForStopSignal = time.Duration(50 * time.Microsecond)
const readMessageSize = 1024

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type GossipPacket struct {
	Simple *SimpleMessage
}

//InitListening allows the start of listening for messages on the server
func InitListening(wg *sync.WaitGroup, connection *net.UDPConn) (chan GossipPacket, chan bool, error) {
	receive := make(chan GossipPacket, 10)
	finish := make(chan bool, 1)

	wg.Add(2)
	go stopListening(finish, connection, wg)
	go listen(receive, finish, connection, wg)
	return receive, finish, nil
}

func stopListening(finish chan bool, connection *net.UDPConn, wg *sync.WaitGroup) {
	<-finish
	connection.Close()
	wg.Done()
	return
}

func listen(receive chan GossipPacket, finish <-chan bool,
	connection *net.UDPConn, wg *sync.WaitGroup) {

	inputBytes := make([]byte, readMessageSize)
	for {
		len, _, err := connection.ReadFrom(inputBytes)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				//indicates that the network was closed and we need to stop listening
				wg.Done()
				return
			}
		}
		if len > 0 {

			var msg GossipPacket
			err = protobuf.Decode(inputBytes, &msg)
			receive <- msg
			inputBytes = make([]byte, readMessageSize)
		}
	}
}

//InitSending allows the start of listening for pings on the server
func InitSending(srcAddress string, dstAddress string, wg *sync.WaitGroup) (chan bool, chan GossipPacket) {
	msgChannel := make(chan GossipPacket, 10)
	finish := make(chan bool, 1)
	wg.Add(1)
	go sendMessage(srcAddress, dstAddress, finish, msgChannel, wg)
	return finish, msgChannel
}

//sendMessage allows the start of sending between a source and a destination
func sendMessage(srcAddress string, dstAddress string, finish <-chan bool,
	msgChannel <-chan GossipPacket, wg *sync.WaitGroup) error {

	destinationAddress, _ := net.ResolveUDPAddr("udp", dstAddress)
	sourceAddress, _ := net.ResolveUDPAddr("udp", srcAddress)

	connection, err := net.DialUDP("udp", sourceAddress, destinationAddress)
	if err != nil {
		wg.Done()
		return err
	}
	for {
		select {
		case <-finish:
			connection.Close()
			wg.Done()
			return nil
		case message := <-msgChannel:
			encoded, err := protobuf.Encode(&message)
			if err != nil {
				connection.Close()
				return err
			}
			connection.Write(encoded)
		}
	}
}
