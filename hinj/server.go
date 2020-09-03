package hinj

import (
	"fmt"
	"log"
	"net"
	"net/url"
)

/*
 * The HINJ server is responsible for:
 *   1. Reading incoming hardware packets
 *   2. Applying modification rules
 *   3. Performing those modifications
 */

type HINJServer struct {
	Addr            net.Addr
	Listener        net.Listener
	shutdownChan    chan int
	shutdownAckChan chan int
}

type URLAddr url.URL

func NewHINJServer(rawURL string) (*HINJServer, error) {
	tmpURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	server := HINJServer{Addr: (*URLAddr)(tmpURL)}
	return &server, nil
}

func (server *HINJServer) Start() error {
	var err error

	server.Listener, err = net.Listen(server.Addr.Network(), server.Addr.String())
	if err != nil {
		return fmt.Errorf("Start(): %s", err)
	}

	// run the server's event loop
	go server.work()

	return nil
}

// Stops the server.
// It is an error to call this function on a server that is not running.
func (server *HINJServer) Shutdown() {
	server.Listener.Close()
	server.shutdownChan <- 0
	<-server.shutdownAckChan
}

func (server *HINJServer) work() {
	keepGoing := true
	for keepGoing {
		conn, err := server.Listener.Accept()
		if err != nil {
			// check if the error was caused by a Shutdown() call
			select {
			case <-server.shutdownChan:
				keepGoing = false
				server.shutdownAckChan <- 1
			default:
				continue
			}
		}
		reader := NewHINJReader(conn)
		writer := NewHINJWriter(conn)
		msg, err := reader.ReadMessage()
		if err != nil {
			log.Printf("HINJServer.work(): error: %s\n", err)
		}

		// TODO: actually process the message
		writer.WriteMessage(msg)
	}
}

// implements net.Addr
func (url *URLAddr) String() string {
	return url.Path
}

// implements net.Addr
func (url *URLAddr) Network() string {
	return url.Scheme
}
