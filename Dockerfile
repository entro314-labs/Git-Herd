# GitHerd Docker Image
FROM alpine:3.20

# Install git (required for GitHerd to function)
RUN apk add --no-cache git ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh githerd

# Copy the binary from GoReleaser build context
COPY githerd /usr/local/bin/githerd

# Ensure binary is executable
RUN chmod +x /usr/local/bin/githerd

# Switch to non-root user
USER githerd

# Set working directory
WORKDIR /workspace

# Add labels for better container management
LABEL org.opencontainers.image.title="GitHerd"
LABEL org.opencontainers.image.description="A concurrent Git repository management tool"
LABEL org.opencontainers.image.source="https://github.com/entro314-labs/Git-Herd"
LABEL org.opencontainers.image.url="https://github.com/entro314-labs/Git-Herd"
LABEL org.opencontainers.image.vendor="entro314-labs"

# Default command
ENTRYPOINT ["/usr/local/bin/githerd"]
CMD ["--help"]