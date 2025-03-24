FROM golang:1.16 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o autoresponderbot ./cmd/autoresponderbot

FROM gcr.io/distroless/base-debian10

WORKDIR /app

COPY --from=builder /app/autoresponderbot /app/autoresponderbot

CMD ["/app/autoresponderbot"]
