package rpc

import (
	"context"
	"hidevops.io/hiboot/pkg/starter/grpc"
	"hidevops.io/mio/node/protobuf"
	"hidevops.io/mio/node/scheduler"
)

type BuildServerImpl struct {
	buildService scheduler.BuildService
}

func newBuildServer(buildService scheduler.BuildService) *BuildServerImpl {
	return &BuildServerImpl{buildService: buildService}
}

func init() {
	//app.Component(newBuildConfigService)
	// must: register grpc server
	// please note that greeterService must implement protobuf.GreeterServer, or it won't be registered.
	grpc.Server(protobuf.RegisterBuildConfigServiceServer, newBuildServer)
}

func (s *BuildServerImpl) SourceCodePull(ctx context.Context, request *protobuf.SourceCodePullRequest) (*protobuf.SourceCodePullResponse, error) {
	go s.buildService.SourceCodePull(request)

	// response to client
	response := &protobuf.SourceCodePullResponse{Code: 200, Message: "OK", Data: request}
	return response, nil
}

func (s *BuildServerImpl) Compile(ctx context.Context, request *protobuf.CompileRequest) (*protobuf.CompileResponse, error) {

	go s.buildService.SourceCodeCompiles(request)

	// response to client
	response := &protobuf.CompileResponse{}
	return response, nil
}

func (s *BuildServerImpl) ImageBuild(ctx context.Context, request *protobuf.ImageBuildRequest) (*protobuf.ImageBuildResponse, error) {

	go s.buildService.SourceCodeImageBuild(request)

	// response to client
	response := &protobuf.ImageBuildResponse{}
	return response, nil
}

func (s *BuildServerImpl) ImagePush(ctx context.Context, request *protobuf.ImagePushRequest) (*protobuf.ImagePushResponse, error) {

	go s.buildService.SourceCodeImagePush(request)

	// response to client
	response := &protobuf.ImagePushResponse{}
	return response, nil
}

func (s *BuildServerImpl) Test(ctx context.Context, request *protobuf.TestsRequest) (*protobuf.TestsResponse, error) {
	go s.buildService.SourceCodeTest(request)

	// response to client
	response := &protobuf.TestsResponse{}
	return response, nil
}

func (s *BuildServerImpl) Command(ctx context.Context, request *protobuf.CommandRequest) (*protobuf.CommandResponse, error) {
	go s.buildService.EnvMakeUp(request)

	// response to client
	response := &protobuf.CommandResponse{}
	return response, nil
}
