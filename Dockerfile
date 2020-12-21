FROM golang:latest

MAINTAINER karokojnr<karoko.jnr99@gmail.com>

#Move to working directory /build
WORKDIR /build

#copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

#Copy the code into the container
COPY . .

EXPOSE 4000

#Build the application
RUN go build -o vepa .

# Build a small image
#FROM scratch
## Copy app to the new image
#COPY --from=builder /build/vepa /
#
## Move config file to new image
#COPY --from=builder /build/docker.env /.env


# Command to run
ENTRYPOINT ["/vepa"]
