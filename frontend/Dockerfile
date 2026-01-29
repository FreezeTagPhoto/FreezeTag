# syntax=docker/dockerfile:1
FROM node:latest AS builder

ARG FREEZETAG_BACKEND_ADDRESS=backend
ENV CI=true

WORKDIR /build

COPY . .
RUN npm install --global pnpm@latest-10
RUN pnpm install --frozen-lockfile
RUN pnpm build

FROM node:current-alpine AS runner

WORKDIR /frontend
COPY --from=builder /build/.next/standalone ./
COPY --from=builder /build/.next/static ./.next/static

ENV NODE_ENV=production
EXPOSE 3000

CMD ["node", "server.js"]
