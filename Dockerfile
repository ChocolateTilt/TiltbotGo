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
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

# Install CA certificates in Alpine Linux
RUN apk --no-cache add ca-certificates && update-ca-certificates

COPY . .
RUN go mod download && \
    GOOS=linux GOARCH=arm go build -o ./app/bot .

############################
# STEP 2 build final image
############################

FROM scratch

ENV TZ=America/Detroit

WORKDIR /app

# Copy the user and group from build
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

# Copy Timezone information
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the CA certificates from the build stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary
COPY --from=build /root/app .

# Use unprivileged user
USER appuser:appuser

CMD [ "./bot" ]