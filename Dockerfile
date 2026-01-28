FROM golang:1.21-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/personaforge .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/personaforge /personaforge

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/personaforge"]


