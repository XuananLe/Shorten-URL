FROM golang:1.22

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .
COPY .env /usr/src/app/

WORKDIR /usr/src/app/cmd/server

# CMD [ "go", "run", "main.go" ]
