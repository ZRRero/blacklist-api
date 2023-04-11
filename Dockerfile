FROM golang:alpine AS build
WORKDIR /build/src/app
COPY . .
RUN go build -o /build/bin/app main.go

FROM alpine
COPY --from=build /build/bin/app /run/bin/app
RUN chmod +x /run/bin/app
ENTRYPOINT ["./run/bin/app"]