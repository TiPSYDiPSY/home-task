ARG GO_VERSION=1.24
ARG APP_VERSION=1.0.0
ARG BUILD_DATE=unknown
ARG VCS_REF=unknown

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN  go build \
    -ldflags "-X main.version=${APP_VERSION} -X main.buildDate=${BUILD_DATE}" \
    -o bin/home-task cmd/home-task/main.go

FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/bin/home-task ./home-task

RUN chown appuser:appgroup home-task && \
    chmod +x home-task

USER appuser

EXPOSE 8080

LABEL org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="your-git-repo-url" \
      org.opencontainers.image.revision="${VCS_REF}"

ENTRYPOINT ["./home-task"]


CMD []

