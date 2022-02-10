# Well Known Types

Protobuf have a set of so-called Well Known Types which provide complex types useful in many situations. These types 
require custom encode and decode functions. This set of types partially implemented in this package.

## [`Empty`](empty.go) [Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Empty)
Empty represent an empty object.

When encode we produce document start and end with nothing in between.

When decode we read document and throw an error if the document contain something.

## [`Duration`](duration.go) [Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#duration)
Represent a span of time.

When decode and encode we use the `protojson`. String field will be produced.

## [`Timestamp`](timestamp.go) [Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#timestamp)
Represent a time point.

When decode and encode we fallback to the default `time.Time` encode/decode function already presented in the `bson`.

Time in `protojson` and `protobson` packages differs from each other.
The `protobson` uses a `DateTime` type from the `mongo-driver` which loose precision but better integrated into the `MongoDB` data model.

> This behavior might be changed in future if it introduce incompatibility or precission loss.

## [`Struct`, `Value`, `ListValue`](struct.go) [Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#struct)
Represent a JSON object.

> When decode and encode we work with int/float/double precision pretty loosely. This might be changed in future if this
> behavior will cause troubles.

## [`Wrappers`](wrappers.go)
[Float Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.FloatValue),
[Double Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.DoubleValue),
[Int32 Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#int32value),
[Int64 Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#int64value),
[UInt32 Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#uint32value),
[UInt64 Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#uint64value),
[Bool Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#boolvalue),
[Bytes Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#bytesvalue),
[String Doc](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#stringvalue)

Represent some regular types.

> When decode and encode we work with int/float/double precision pretty loosely. This might be changed in future if this
> behavior will cause troubles.
