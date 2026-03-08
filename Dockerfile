# Build frontend
FROM node:latest AS build-frontend
WORKDIR /usr/src/app
COPY ./frontend/ ./
RUN npm install
RUN npm run build

# Build backend
FROM golang:latest AS build-backend
WORKDIR /usr/src/app/server
COPY ./server/go.* ./
RUN go mod download
COPY ./server .
RUN CGO_ENABLED=0 go build -o main ./cmd/api

# Final image: backend + frontend static files
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build-backend /usr/src/app/server/main .
COPY --from=build-backend /usr/src/app/server/migrations ./migrations
COPY --from=build-frontend /usr/src/app/dist ./frontend
ENV STATIC_DIR=/app/frontend
ENV SERVER_PORT=8080
EXPOSE 8080
CMD ["./main"]
