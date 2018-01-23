# Sample go server that uses the Grafeas go client library

To run this server against the Grafeas reference implementation run the following:

```
go get github.com/Grafeas/Greafeas/samples/server/go-server/api/
go get github.com/Grafeas/client-go
go run grafeas/samples/server/go-server/api/server/main/main.go
go run client-go/example/v1alpha1/main.go
