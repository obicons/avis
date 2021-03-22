//go:generate protoc -I=./ --go_out=../ --go-grpc_out=../ ./simulator_controller.proto
package controller

import (
	context "context"
	"net"
	"net/url"

	"github.com/obicons/avis/sim"
	"google.golang.org/grpc"
)

type SimulatorController struct {
	url        *url.URL
	grpcServer *grpc.Server
	simulator  sim.Sim
	listener   net.Listener
	shutdownCh chan int
	modeCh     chan int
}

func New(addrStr string, simulator sim.Sim) (*SimulatorController, error) {
	var err error
	server := SimulatorController{}
	server.url, err = url.Parse(addrStr)
	if err != nil {
		return nil, err
	}
	server.simulator = simulator
	server.shutdownCh = make(chan int)
	server.modeCh = make(chan int)
	return &server, nil

}

// Starts the SimulatorController.
// It is an error to call this method if server has already been started.
func (server *SimulatorController) Start() error {
	var err error
	// TODO -- clean this up
	server.listener, err = net.Listen(server.url.Scheme, server.url.Path)
	if err != nil {
		return err
	}
	server.grpcServer = grpc.NewServer()
	service := NewSimulatorControllerService(server)
	RegisterSimulatorControllerService(server.grpcServer, service)
	return server.grpcServer.Serve(server.listener)
}

// Returns a channel to receive mode changes from
func (server *SimulatorController) Mode() <-chan int {
	return server.modeCh
}

// Returns a channel to receieve shutdown requests from.
func (server *SimulatorController) Done() <-chan int {
	return server.shutdownCh
}

// Stops the SimulatorController.
// It is an error to call this method if server has not been started.
func (server *SimulatorController) Shutdown() {
	server.grpcServer.Stop()
	server.listener.Close()
}

// Implements RPC
func (s *SimulatorController) Step(ctx context.Context, req *StepRequest) (*StepResponse, error) {
	err := s.simulator.Step(ctx)
	return &StepResponse{}, err
}

// Implements RPC
func (s *SimulatorController) Position(ctx context.Context, req *PositionRequest) (*PositionResponse, error) {
	position, err := s.simulator.Position(ctx)
	return &PositionResponse{X: position.X, Y: position.Y, Z: position.Z}, err
}

// Implements RPC
func (s *SimulatorController) Time(ctx context.Context, req *TimeRequest) (*TimeResponse, error) {
	time, err := s.simulator.SimTime(ctx)
	return &TimeResponse{TvSec: uint64(time.Second()), TvUSec: uint64(1000 * time.Nanosecond())}, err
}

// Implements RPC
func (s *SimulatorController) Terminate(ctx context.Context, req *TerminateRequest) (*TerminateResponse, error) {
	s.shutdownCh <- 1
	return &TerminateResponse{}, nil
}

// Implements RPC
func (s *SimulatorController) ModeChange(ctx context.Context, req *ModeChangeRequest) (*ModeChangeResponse, error) {
	s.modeCh <- int(req.NextMode)
	return &ModeChangeResponse{}, nil
}
