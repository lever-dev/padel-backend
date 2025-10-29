FROM golang:1.25

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0 \
    GOFLAGS=-mod=vendor

CMD ["bash"]
