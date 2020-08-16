## How to contribute
This project required a go 1.11 (or greater) version. It's written using the gomodule approach.

```
#Installing mockgen depencendy
GO111MODULE=on go get github.com/golang/mock/mockgen@v1.4.4

# seting up joal
git clone git@github.com:owl-torrent/joal-cli.git
cd joal-cli
go generate ./...
```

This project use mockgen, when modifying interface/struct declaration that are target of mockgen (check for //go:generate mockgen ... [XXX,XXX] at the top of the file) you must regenerate the mock implementation with `go generate` using:
- `go generate ./...` from the root of the project to regenerate all the mocks
- `go generate filename.go` to process only the mockgen declared in this file
