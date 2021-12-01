# link-tracking

This is my first Golang project created to try Golang, gRPC, OpenAPI and be able to serve my own shortened links and tracking pixels.

Application that generates shortened links for given URL and keeps info about user and other metadata. This application also listens for requests to the generated URLs and returns the original URL.

We also generate links for link hits that can be used as tracking pixels.

The application reachable through gRPC and OpenAPI thanks to OpenAPI gateway.

1. gRPC endpoint is running on https://0.0.0.0:10000/.
2. OpenAPI endpoint is running on https://0.0.0.0:3000/.
3. An OpenAPI UI is served on https://0.0.0.0:11000/.

To make this operational:
1. Install `buf` with `make install`, which is necessary for us to generate the Go and OpenAPIv2 files.
2. generate the protobufs with `make generate`.
3. Now you can run the web server with `go run main.go`.


There is also dockerfile prepared:

```
docker build -t golinks .
docker run -e DB_PATH="sqlite.db" -e AUTH_PASSWORD="pass" -p 11000:11000 -p 10000:10000 -p 3000:3000 golinks
```