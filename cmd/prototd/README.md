# prototd
A simple command line tool to generate a Thing Description from Protocol Buffers

prototd takes one `.proto` file that contains one gRPC serivce with RPCs (currently only Unary RPCs) and the Messages used by them, and generate a W3C Web of Things Thing Description for the gRPC service that a HTTP2 client can consume.

Following [the document about an imlementation of gRPC](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) carried over HTTP2 framing, we generate the corresponding [Interaction Affordances](https://www.w3.org/TR/wot-thing-description/#interactionaffordance) to the RPCs exposed by the gRPC service.

## Mapping from Protocol Buffers to Thing Description

protod takes the IP address and the port where the gRPC service is hosted to provide the target IRI in the [`forms`](https://www.w3.org/TR/wot-thing-description/#form).

protod implements a policy to classify RPCs into one of the Interaction Affordances: Property, Action, and Event.

This classification assumes that the provided protobuf file is conformal [the Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style) as well as other generic naming conventions, such as `GetPropertyName` for accessing a property `PropertyName`.

For encoding unary RPCs, those functions which do not take input parameters are assumed to take an `Empty` message such as:
```proto
message Empty {
}
```

The policy is encoded in [the handlers](https://pkg.go.dev/github.com/emicklei/proto@v1.9.2#Handler) by parsing the protobuf file with [`github.com/emicklei/proto`](https://github.com/emicklei/proto).

prototd uses the `ThingDescription` type from [`github.com/linksmart/thing-directory/wot`](https://github.com/linksmart/thing-directory/blob/master/wot/thing_description.go) for JSON marshaller.

## Usage

```console
NAME:
   prototd - Translate ProtocolBuffers to ThingDescription

USAGE:
   prototd [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value, -p value  The port for the gRPC service (default: 50051)
   --ip value              The IP address for the gRPC serivce (default: "127.0.0.1")
   --output FILE, -o FILE  Write the resulting Thing Description to FILE (default: "td.jsonld")
   --help, -h              show help (default: false)
```
