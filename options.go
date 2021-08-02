package cloudrunner

import (
	"go.einride.tech/cloudrunner/cloudconfig"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Option provides optional configuration for a run context.
type Option func(*runContext)

// WithRequestLoggerMessageTransformer configures the request logger with a message transformer.
func WithRequestLoggerMessageTransformer(transformer func(proto.Message) proto.Message) Option {
	return func(run *runContext) {
		run.requestLoggerMiddleware.MessageTransformer = transformer
	}
}

// WithConfig configures an additional config struct to be loaded.
func WithConfig(name string, config interface{}) Option {
	return func(run *runContext) {
		run.configOptions = append(run.configOptions, cloudconfig.WithAdditionalSpec(name, config))
	}
}

// WithOptions configures the run context with a list of options.
func WithOptions(options []Option) Option {
	return func(run *runContext) {
		for _, option := range options {
			option(run)
		}
	}
}

// WithGRPCServerOptions configures the run context with additional default options for NewGRPCServer.
func WithGRPCServerOptions(grpcServerOptions []grpc.ServerOption) Option {
	return func(run *runContext) {
		run.grpcServerOptions = append(run.grpcServerOptions, grpcServerOptions...)
	}
}
