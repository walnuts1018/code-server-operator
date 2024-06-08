FROM golang:1.22 as builder
ENV ROOT=/build
RUN mkdir ${ROOT}
WORKDIR ${ROOT}

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main $ROOT/main.go && chmod +x ./main

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && \
    apt-get install -y curl && \ 
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/main ./
COPY --from=builder /usr/share/zoneinfo/Asia/Tokyo /usr/share/zoneinfo/Asia/Tokyo
CMD ["./main"]
