all: ./test.pb.go

current_mkfile:=$(abspath $(lastword $(MAKEFILE_LIST)))
current_dir:=$(patsubst %/,%,$(dir $(current_mkfile)))

current_dir_name:=$(notdir $(current_dir))

./test.pb.go:
	protoc -I $(current_dir) \
			test.proto \
			--go_out=. --go_opt=module=github.com/rdmrcv/go-protom/pkg/internal
