FROM golang:1.26.5 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./agent

FROM gcr.io/distroless/static-debian12
COPY --from=build /agent /agent
ENTRYPOINT ["/agent"]