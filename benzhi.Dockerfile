FROM golang:1.26.2
WORKDIR /app
ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor ./...
CMD ["go", "run", "-mod=vendor", "./cmd/launchguard"]
