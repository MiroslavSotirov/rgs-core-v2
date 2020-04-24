# multi stage docker build
FROM golang:1.13 AS builder
COPY . /go/src/rgs-core-v2
WORKDIR /go/src/rgs-core-v2
RUN go mod download
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -a -o /rgs ./cmd

# final stage
FROM alpine:3.10
RUN apk --no-cache add ca-certificates
# copy rgs binary
COPY --from=builder /rgs ./
#copy playcheck template
COPY --from=builder /go/src/rgs-core-v2/templates ./templates
RUN chmod +x ./rgs
ENTRYPOINT ["./rgs", "-logtostderr=true", "-gethashes=false"]
EXPOSE 3000
