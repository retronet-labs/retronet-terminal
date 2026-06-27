FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY . .

RUN go build -o /out/retronet-terminal ./cmd/retronet-terminal

FROM alpine:latest

WORKDIR /app
COPY --from=builder /out/retronet-terminal /app/retronet-terminal

ENTRYPOINT ["/app/retronet-terminal"]
CMD ["-demo", "-screen"]
