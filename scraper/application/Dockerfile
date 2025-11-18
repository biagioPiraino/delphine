FROM golang:1.24-alpine AS builder

# set env variable during building
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG ARCH=amd64

# working directoy inside container
WORKDIR /build

# copy module and sum files for dependency installation
COPY go.mod go.sum ./

# download dependencies
RUN go mod download

# copy the app
COPY . .

# build the binary
RUN go build -o /app .

FROM alpine:3.21 AS final

# set env variables for container
ENV KAFKA_PORT=9094

# copy app to binary
COPY --from=builder /app /bin/app

# expose port
EXPOSE 8080

# run
CMD ["bin/app"]