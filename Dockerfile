# Build stage
FROM golang AS build-env
ADD . /src/grpc-gateway-boilerplate
ENV CGO_ENABLED=1
RUN cd /src/grpc-gateway-boilerplate && go build -o /app

# Production stage
FROM debian
RUN mkdir /server
COPY --from=build-env /app /server/app

COPY ./sql /server/sql/
COPY ./third_party /server/third_party/
ENTRYPOINT ["/server/app"]
