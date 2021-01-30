#!/usr/bin/sh

rm -rf pkg/mock

mockgen \
  -source=pkg/std/user/user.go \
  -destination=pkg/mock/std/user/user.go

mockgen \
  -source=pkg/std/runtime/runtime.go \
  -destination=pkg/mock/std/runtime/runtime.go

mockgen \
  -source=pkg/std/os/os.go \
  -destination=pkg/mock/std/os/os.go

mockgen \
  -source=pkg/std/ioutil/ioutil.go \
  -destination=pkg/mock/std/ioutil/ioutil.go
