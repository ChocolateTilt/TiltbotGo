# syntax=docker/dockerfile:1
FROM golang:alpine

WORKDIR /app

RUN apk add tzdata

ENV TZ=America/Detroit

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
    GOOS=linux GOARCH=arm go build -o ./bot .

CMD [ "./bot" ]