# shortfree: MCP server and CLI for free 1-4 character domains.
#
#   docker run -i --rm ghcr.io/maximilianfeix/shortfree            # MCP server on stdio
#   docker run --rm ghcr.io/maximilianfeix/shortfree check q7.tools  # CLI
FROM --platform=$BUILDPLATFORM golang:1.27 AS build
ARG TARGETOS TARGETARCH VERSION=""
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/shortfree ./cmd/shortfree

FROM gcr.io/distroless/static-debian12:nonroot
LABEL io.modelcontextprotocol.server.name="io.github.maximilianfeix/shortfree" \
      org.opencontainers.image.title="shortfree" \
      org.opencontainers.image.description="MCP server and CLI: live domain checks and every free 1-4 character domain on .com .net .de .ai .io .dev .app .tools" \
      org.opencontainers.image.source="https://github.com/maximilianfeix/shortfree" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /out/shortfree /shortfree
USER nonroot
ENTRYPOINT ["/shortfree"]
CMD ["mcp"]
