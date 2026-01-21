# syntax=docker/dockerfile:1
FROM node:latest

COPY src /frontend/src
COPY next.config.ts /frontend/next.config.ts
COPY package.json /frontend/package.json
COPY pnpm-lock.yaml /frontend/pnpm-lock.yaml
COPY pnpm-workspace.yaml /frontend/pnpm-workspace.yaml
COPY tsconfig.json /frontend/tsconfig.json

WORKDIR /frontend
RUN npm install --global pnpm@latest-10
RUN mkdir -p .pnpm-store
RUN pnpm config set store-dir .pnpm-store
RUN pnpm install
RUN pnpm build

EXPOSE 3000

CMD ["pnpm",  "start"]