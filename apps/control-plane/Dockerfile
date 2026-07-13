FROM golang:1.22-alpine AS build

WORKDIR /src
COPY apps/control-plane/go.mod apps/control-plane/go.sum ./
RUN go mod download

COPY apps/control-plane/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/qualora-api .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/qualora-api /app/qualora-api

EXPOSE 8080
CMD ["/app/qualora-api"]
