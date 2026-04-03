# build stage
FROM golang:1.26.1 AS builder
WORKDIR /app

# copy rest of code
COPY server/ ./
# COPY ./server/ .
RUN go build -o app

# run stage
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/config ./config 
CMD ["./app"]