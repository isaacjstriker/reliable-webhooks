# Multi-stage build for minimal runtime image
FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO-ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/server /app/server
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT [ "/app/server" ]