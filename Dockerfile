FROM golang:1.25-alpine

WORKDIR /app

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum* ./

RUN if [ -f go.mod ]; then go mod download; fi

COPY . .

CMD ["air", "-c", ".air.toml"]
