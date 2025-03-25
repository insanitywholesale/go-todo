# build stage
FROM golang:1.24 AS build

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/go-todo
COPY . .

RUN go get -v
RUN go vet -v
RUN go install -v

# run stage
FROM gcr.io/distroless/static-debian12:nonroot AS run

COPY --from=build /go/bin/go-todo /

EXPOSE 8080

CMD ["/go-todo"]
