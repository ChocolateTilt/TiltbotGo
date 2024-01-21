# syntax=docker/dockerfile:1

################################
# STEP 1 build executable binary
################################

FROM golang:alpine as build

WORKDIR /root

RUN apk add tzdata

# Create appuser.
ENV USER=appuser
ENV UID=10001 
RUN apk add tzdata --no-cache && \
    adduser --disabled-password --gecos "" --home "/nonexistent" --shell "/sbin/nologin" --no-create-home --uid "${UID}" "${USER}" && \
    apk --no-cache add ca-certificates && update-ca-certificates

COPY . .
RUN go mod download && \
    go build -o ./app/bot .

############################
# STEP 2 build final image
############################

FROM scratch

ENV TZ=America/Detroit

WORKDIR /app

# Copy the user, group, timezone information, CA certificates, and binary from the build stage
COPY --from=build /etc/passwd /etc/passwd \
    /etc/group /etc/group \
    /usr/share/zoneinfo /usr/share/zoneinfo \
    /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt \
    /root/app /app/

# Use unprivileged user
USER appuser:appuser

CMD [ "./bot" ]