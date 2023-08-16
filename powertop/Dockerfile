FROM golang:latest as builder

# Copy Telegraf plugin directory to the image
WORKDIR /telegraf 
COPY /telegraf .

# Build the command inside the container
RUN make

FROM debian:buster-slim

# Copy the Telegraf binary
COPY --from=builder /telegraf/telegraf /usr/bin/

# Copy over the plugin configuration
COPY telegraf.conf /etc/telegraf/

# Run Telegraf when the container starts
ENTRYPOINT ["telegraf"]
