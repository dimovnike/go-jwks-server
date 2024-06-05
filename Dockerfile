###############  builder stage ###################

FROM golang:1.21.11-alpine3.20 as builder

WORKDIR /build/

COPY go.mod go.sum ./
RUN go mod download

# add the rest of the code and build the app
ADD . ./
RUN go build -o /jwks-server cmd/main.go

################ final stage #########################

FROM alpine:3.20

COPY --from=builder /jwks-server /usr/local/bin/

CMD ["/usr/local/bin/jwks-server"]
