# Development/runtime container for the Cloudflare Workers (TypeScript) app.
#
# This image runs the Worker locally with `wrangler dev` (workerd) inside the
# container. It is useful for running the TS backend on Docker hosts that are
# not the Cloudflare network. Production deploys still go through
# `wrangler deploy` / GitHub Actions.
#
# Use a glibc-based image: Cloudflare's workerd-linux-64 binary is not
# compatible with musl-based images such as node:alpine.
FROM node:24-slim

# CA certificates are required for workerd's outbound HTTPS fetches.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install dependencies first so they are cached separately.
COPY package.json package-lock.json ./
RUN npm ci

# Copy the rest of the project.
COPY . .

ENV NODE_ENV=production
ENV WRANGLER_SEND_METRICS=false
# Force non-interactive Wrangler prompts (skip D1 migration confirmation).
ENV CI=true

EXPOSE 8080

# Local D1 state lives under /app/.wrangler; mount a volume there to persist
# the SQLite cache between container restarts.
CMD ["sh", "-c", "npx wrangler d1 migrations apply stock-fetcher --local < /dev/null && exec npx wrangler dev --local --ip 0.0.0.0 --port 8080"]
