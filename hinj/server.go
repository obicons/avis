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
	gyroReadings             int
	accelReadings            int
	gpsReadings              int
	compassReadings          int
	baroReadings             int
}

type URLAddr url.URL

func NewHINJServer(rawURL string) (*HINJServer, error) {
	tmpURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	server := HINJServer{
		Addr: (*URLAddr)(tmpURL),

		// this needs to have a buffer to avoid deadlock
		shutdownChan: make(chan int, 1),

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
// It is alright to recycle the server after calling this function.
func (server *HINJServer) Shutdown() {
	server.shutdownChan <- 0
	server.Listener.Close()
	<-server.shutdownAckChan
	server.reportStats()
	server.resetFailures()
}

// Causes all future reads of the provided sensor category and instance to fail.
func (server *HINJServer) FailSensor(sensorType Sensor, instanceNo uint8) {
	if server.failureStateBySensorType[sensorType] == nil {
		server.failureStateBySensorType[sensorType] = make(map[uint8]bool)
	}
	server.failureStateBySensorType[sensorType][instanceNo] = true
}

// Resets the failure state
func (server *HINJServer) resetFailures() {
	server.failureStateBySensorType = make(map[Sensor]map[uint8]bool)
}

func (server *HINJServer) recordStats(msg interface{}) {
	switch msg.(type) {
	case *GPSPacket:
		server.gpsReadings++
	case *AccelerometerPacket:
		server.accelReadings++
	case *GyroscopePacket:
		server.gyroReadings++
	case *BarometerPacket:
		server.baroReadings++
	case *CompassPacket:
		server.compassReadings++
	}
}

func (server *HINJServer) reportStats() {
	fmt.Printf("GPS readings: %d\n", server.gpsReadings)
	fmt.Printf("Accel readings: %d\n", server.accelReadings)
	fmt.Printf("Gyro readings: %d\n", server.gyroReadings)
	fmt.Printf("Compass readings: %d\n", server.compassReadings)
	fmt.Printf("Baro readings: %d\n", server.baroReadings)
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
				continue
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

		server.recordStats(msg)

		server.checkAndFail(msg)
		err = writer.WriteMessage(msg)
		if err != nil {
			log.Printf("HINJServer.work(): error writing: %s\n", err)
		}
		conn.Close()
	}
	server.shutdownAckChan <- 0
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
