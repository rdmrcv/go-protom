# Go Proto MongoDB Utils

The lib is not considered as stable and an API might be broken until `v1.x.x` version is tagged.

This lib provides some utils to work with protobuf messages together with MongoDB.

Main reason to create this package is a [`protojson`](https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson)
package that encodes `proto` messages differently than an `encodying/json` package. This creates a big gap between
regular golang structs and structs provided by a `protoc`.

## [`Protobson`](pkg/protobson)
### Message
This package is highly inspired by the original `protojson` package but bound to the `go.mongodb.org/mongo-driver` 
package.

The package provides codecs and a registry that compatible with the `mongo-dirver` and fits well into the `mongo-driver` 
encode/decode process. The codec designed mostly as intermediate component who can encode and decode the `proto.Message`
and leverage the `mongo-driver` registry machinery to encode or decode primitive types and lists/maps.

### Enum
Only special case is enums. Enums cannot be properly decoded because in `protoreflect` they represented as 
the `protoreflect.EnumNumber` which completely loose an original enum type, so when we encode or decode an enum value 
represented as string we cannot guess which enum we currently encode or decode.

To solve this we have two ways:

1. Embed the enum codec into the `ProtoMessageCodec`.
2. Implement the `ProtoEnumCodec` and introduce special handling of the enum values in the `ProtoMessageCodec`

In the protojson package implemented second solution. It decouple encode and decode of enums from
the `ProtoMessageCodec`. But still introduce pretty ugly way to operate with enum fields in the message.

## Usage:

If you use a default registry from `mongo-driver` you can just overwrite it in client — registry provided by
the `protobson.BsonPBRegistry` should be equal to default one, just with special hooks that will handle protobuf models.

If you customize your registry — you should read the [`registry.go`](pkg/protobson/registry.go) file and see how you can
add required codecs into registry.

## [Well Known Types](pkg/protobson/types)

Partially implemented with some caveats. To see details - go to README in `protobson/types` package.
