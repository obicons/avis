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
	Addr                     net.Addr
	Listener                 net.Listener
	shutdownChan             chan int
	shutdownAckChan          chan int
	failureStateBySensorType map[Sensor]map[uint8]bool
	enableFailureChan        chan SensorFailure
}

type URLAddr url.URL

func NewHINJServer(rawURL string) (*HINJServer, error) {
	tmpURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	server := HINJServer{
		Addr:                     (*URLAddr)(tmpURL),
		shutdownChan:             make(chan int),
		shutdownAckChan:          make(chan int),
		enableFailureChan:        make(chan SensorFailure),
		failureStateBySensorType: make(map[Sensor]map[uint8]bool),
	}

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

// Causes all future reads of the provided sensor category and instance to fail.
func (server *HINJServer) FailSensor(sensorType Sensor, instanceNo uint8) {

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
				log.Printf("HINJServer.work(): error %s\n", err)
				continue
			}
		}
		server.checkForPendingFailures()
		reader := NewHINJReader(conn)
		writer := NewHINJWriter(conn)
		msg, err := reader.ReadMessage()
		if err != nil {
			log.Printf("HINJServer.work(): error: %s\n", err)
		}

		server.checkAndFail(msg)
		err = writer.WriteMessage(msg)
		if err != nil {
			log.Printf("HINJServer.work(): error writing: %s\n", err)
		}
		conn.Close()
	}
}

func (server *HINJServer) checkForPendingFailures() {
	keepGoing := true
	for keepGoing {
		select {
		case failure := <-server.enableFailureChan:
			server.failureStateBySensorType[failure.SensorType][failure.Instance] = true
		default:
			keepGoing = false
		}
	}
}

func (server *HINJServer) checkAndFail(msg interface{}) {
	if gpsPacket, ok := msg.(*GPSPacket); ok {
		if server.failureStateBySensorType[GPS][gpsPacket.Instance] {
			gpsPacket.Ignore = 1
		}
	} else if accelPacket, ok := msg.(*AccelerometerPacket); ok {
		if server.failureStateBySensorType[Accelerometer][accelPacket.Instance] {
			accelPacket.Ignore = 1
		}
	} else if gyroPacket, ok := msg.(*GyroscopePacket); ok {
		if server.failureStateBySensorType[Gyroscope][gyroPacket.Instance] {
			gyroPacket.Ignore = 1
		}
	} else if baroPacket, ok := msg.(*BarometerPacket); ok {
		if server.failureStateBySensorType[Barometer][baroPacket.Instance] {
			baroPacket.Ignore = 1
		}
	} else if compassPacket, ok := msg.(*CompassPacket); ok {
		if server.failureStateBySensorType[Compass][compassPacket.Instance] {
			compassPacket.Ignore = 1
		}
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
