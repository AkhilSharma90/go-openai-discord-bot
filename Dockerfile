FROM golang:1.20-alpine

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-openai-bot-discord

CMD [ "/go-openai-bot-discord" ]
