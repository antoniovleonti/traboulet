FROM golang:1.19
COPY ./src ./
WORKDIR ./main
RUN CGO_ENABLED=0 GOOS=linux go build -o main
CMD "./main"
