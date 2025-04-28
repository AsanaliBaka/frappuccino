FROM golang:1.23

WORKDIR /app

COPY . .

COPY wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]