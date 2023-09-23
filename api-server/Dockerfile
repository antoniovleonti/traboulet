FROM golang:1.21.0
COPY ./src ./
WORKDIR ./main
RUN CGO_ENABLED=0 GOOS=linux go build -o main
CMD "./main"
