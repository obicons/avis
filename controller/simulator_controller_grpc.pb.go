// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package controller

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// SimulatorControllerClient is the client API for SimulatorController service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SimulatorControllerClient interface {
	Step(ctx context.Context, in *StepRequest, opts ...grpc.CallOption) (*StepResponse, error)
	Position(ctx context.Context, in *PositionRequest, opts ...grpc.CallOption) (*PositionResponse, error)
}

type simulatorControllerClient struct {
	cc grpc.ClientConnInterface
}

func NewSimulatorControllerClient(cc grpc.ClientConnInterface) SimulatorControllerClient {
	return &simulatorControllerClient{cc}
}

var simulatorControllerStepStreamDesc = &grpc.StreamDesc{
	StreamName: "Step",
}

func (c *simulatorControllerClient) Step(ctx context.Context, in *StepRequest, opts ...grpc.CallOption) (*StepResponse, error) {
	out := new(StepResponse)
	err := c.cc.Invoke(ctx, "/controller.SimulatorController/Step", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var simulatorControllerPositionStreamDesc = &grpc.StreamDesc{
	StreamName: "Position",
}

func (c *simulatorControllerClient) Position(ctx context.Context, in *PositionRequest, opts ...grpc.CallOption) (*PositionResponse, error) {
	out := new(PositionResponse)
	err := c.cc.Invoke(ctx, "/controller.SimulatorController/Position", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SimulatorControllerService is the service API for SimulatorController service.
// Fields should be assigned to their respective handler implementations only before
// RegisterSimulatorControllerService is called.  Any unassigned fields will result in the
// handler for that method returning an Unimplemented error.
type SimulatorControllerService struct {
	Step     func(context.Context, *StepRequest) (*StepResponse, error)
	Position func(context.Context, *PositionRequest) (*PositionResponse, error)
}

func (s *SimulatorControllerService) step(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StepRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.Step(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     s,
		FullMethod: "/controller.SimulatorController/Step",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.Step(ctx, req.(*StepRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func (s *SimulatorControllerService) position(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PositionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.Position(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     s,
		FullMethod: "/controller.SimulatorController/Position",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.Position(ctx, req.(*PositionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RegisterSimulatorControllerService registers a service implementation with a gRPC server.
func RegisterSimulatorControllerService(s grpc.ServiceRegistrar, srv *SimulatorControllerService) {
	srvCopy := *srv
	if srvCopy.Step == nil {
		srvCopy.Step = func(context.Context, *StepRequest) (*StepResponse, error) {
			return nil, status.Errorf(codes.Unimplemented, "method Step not implemented")
		}
	}
	if srvCopy.Position == nil {
		srvCopy.Position = func(context.Context, *PositionRequest) (*PositionResponse, error) {
			return nil, status.Errorf(codes.Unimplemented, "method Position not implemented")
		}
	}
	sd := grpc.ServiceDesc{
		ServiceName: "controller.SimulatorController",
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Step",
				Handler:    srvCopy.step,
			},
			{
				MethodName: "Position",
				Handler:    srvCopy.position,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "simulator_controller.proto",
	}

	s.RegisterService(&sd, nil)
}

// NewSimulatorControllerService creates a new SimulatorControllerService containing the
// implemented methods of the SimulatorController service in s.  Any unimplemented
// methods will result in the gRPC server returning an UNIMPLEMENTED status to the client.
// This includes situations where the method handler is misspelled or has the wrong
// signature.  For this reason, this function should be used with great care and
// is not recommended to be used by most users.
func NewSimulatorControllerService(s interface{}) *SimulatorControllerService {
	ns := &SimulatorControllerService{}
	if h, ok := s.(interface {
		Step(context.Context, *StepRequest) (*StepResponse, error)
	}); ok {
		ns.Step = h.Step
	}
	if h, ok := s.(interface {
		Position(context.Context, *PositionRequest) (*PositionResponse, error)
	}); ok {
		ns.Position = h.Position
	}
	return ns
}

// UnstableSimulatorControllerService is the service API for SimulatorController service.
// New methods may be added to this interface if they are added to the service
// definition, which is not a backward-compatible change.  For this reason,
// use of this type is not recommended.
type UnstableSimulatorControllerService interface {
	Step(context.Context, *StepRequest) (*StepResponse, error)
	Position(context.Context, *PositionRequest) (*PositionResponse, error)
}
