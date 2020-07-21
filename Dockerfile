# Telling to use Docker's golang ready image
FROM golang
# Name and Email of the author
MAINTAINER Karoko JNR <karoko.jnr99@gmail.com>
# Create app folder
RUN mkdir /vepa
# Copy our file in the host contianer to our contianer
ADD . /vepa
# Set /app to the go folder as workdir
WORKDIR /vepa
# Generate binary file from our /app
RUN go build
# Expose the port 4000
EXPOSE 4000:4000
# Run the app binarry file
CMD ["./vepa"]