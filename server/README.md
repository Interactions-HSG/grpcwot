# prototd-server
A simple server to receive a proto file parsed into interaction affordances and data schemas

The server takes one `.proto` file that contains one gRPC serivce with RPCs (currently only Unary RPCs) and the Messages used by them, 
and returns interaction affordances with data schemes for their request and response type from the RPCs to allow
users an easier classification over a user-friendly frontend

## Mapping from Protocol Buffers to Thing Description

prototd-server implements a policy to pre-classify RPCs into one of the Interaction Affordances: Property, Action, and Event.

This classification assumes that the provided protobuf file is conformal [the Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style) as well as other generic naming conventions, such as `GetPropertyName` for accessing a property `PropertyName`.

For encoding unary RPCs, those functions which do not take input parameters are assumed to take an `Empty` message such as:
```proto
message Empty {
}
```

The policy is encoded in [the handlers](https://pkg.go.dev/github.com/emicklei/proto@v1.9.2#Handler) by parsing the protobuf file with [`github.com/emicklei/proto`](https://github.com/emicklei/proto).


## Usage

prototd-server is exposed on port 8080 and the .proto file can be sent to the endpoint "/upload".
A prototyped HTTP request looks as the following: 
```
POST /upload HTTP/1.1
Host: localhost:8080
Content-Length: 240
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW

----WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="uploadfile"; filename="<fileSrc>"
Content-Type: <Content-Type header here>

(data)
----WebKitFormBoundary7MA4YWxkTrZu0gW

```

The response is a json formatted representation of RPCs, classified into interaction affordance categories,
together with their request and response as DataSchemas.
The response follows the following pattern:

```json
{
  "Props": [
    {
      "Name": "PropertyName",
      "GetProp": {
        "Name": "Name of the corresponding Get Property",
        "Req": {
          "Type": "type of the request data schema",
          "Properties": "properties of the request (could be again a map of dataSchemes)"
        },
        "Res": {
          "Type": "type of the response data schema",
          "Properties": "properties of the response (could be again a map of dataSchemes)"
        }
      },
      "SetProp": {
        "Name": "Name of the corresponding Get Property",
        "Req": {
          "Type": "type of the request data schema",
          "Properties": "properties of the request (could be again a map of dataSchemes)"
        },
        "Res": {
          "Type": "type of the response data schema",
          "Properties": "properties of the response (could be again a map of dataSchemes)"
        }
      },
      "Category": "int 0-2 [0: readonly, 1: writeonly, 2: readwrite]"
    }
  ],
  "Actions": [
    {
      "Name": "Name of a RPC characterized as action",
      "Req": {
        "Type": "type of the request data schema",
        "Properties": "properties of the request (could be again a map of dataSchemes)"
      },
      "Res": {
        "Type": "type of the response data schema",
        "Properties": "properties of the response (could be again a map of dataSchemes)"
      }
    }
  ],
  "Events": [
    {
      "Name": "Name of a RPC characterized as event",
      "Req": {
        "Type": "type of the request data schema",
        "Properties": "properties of the request (could be again a map of dataSchemes)"
      },
      "Res": {
        "Type": "type of the response data schema",
        "Properties": "properties of the response (could be again a map of dataSchemes)"
      }
    }
  ]
}
```

The concrete building interface for this is:

```go
type serverAffordances struct {
	Props   []serverProperty
	Actions []serverAffordance
	Events  []serverAffordance
}

type serverProperty struct {
	Name     string
	GetProp  serverAffordance
	SetProp  serverAffordance
	Category int
}

type serverAffordance struct {
	Name string
	Req  serverDataSchema
	Res  serverDataSchema
}

type serverDataSchema struct {
	Type       string
	Properties []serverProp
}

type serverProp struct {
	Key   string
	Value serverDataSchema
}
```