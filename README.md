# Go Proto MongoDB Utils

The lib is not considered as stable and an API might be broken until `v1.x.x` version is tagged.

This lib provides some utils to work with protobuf messages together with MongoDB.

Main reason to create this package is a [`protojson`](https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson)
package that encodes `proto` messages differently than an `encodying/json` package. This creates a big gap between
ordinal golang structs and structs provided by a `protoc`.

The lib solve that gap by provide two packages:

## [`Protobson`](pkg/protobson)
This part is highly inspired by the original `protojson` package.

Also, I take a look at the great [`romnn/bsonpb`](https://github.com/romnn/bsonpb) package, but it is requires Bazel
which I not use and not so tightly integrated with the MongoDB.

## [`Registry`](pkg/registry)
This package provides the registry that register hooks which will detect `proto.Message` and decode/encode them with a
codec provided via the `protobson` package.
