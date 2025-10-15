# git-herd Docker Image
FROM alpine:3.20

# Install git (required for git-herd to function)
RUN apk add --no-cache git ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh git-herd

# Copy the binary from GoReleaser build context
COPY git-herd /usr/local/bin/git-herd

# Ensure binary is executable
RUN chmod +x /usr/local/bin/git-herd

# Switch to non-root user
USER git-herd

# Set working directory
WORKDIR /workspace

# Add labels for better container management
LABEL org.opencontainers.image.title="git-herd"
LABEL org.opencontainers.image.description="A concurrent Git repository management tool"
LABEL org.opencontainers.image.source="https://github.com/entro314-labs/git-herd"
LABEL org.opencontainers.image.url="https://github.com/entro314-labs/git-herd"
LABEL org.opencontainers.image.vendor="entro314-labs"

# Default command
ENTRYPOINT ["/usr/local/bin/git-herd"]
CMD ["--help"]