# Web app image for local/self-hosted runs (production deploys go to Vercel).
# Build from the repo root: docker compose -f infra/docker-compose.yaml --profile app build
FROM oven/bun:1.4.2 AS deps
WORKDIR /app
COPY package.json bun.lock ./
COPY apps/web/package.json apps/web/
# No .git inside the image: skip the husky `prepare` hook.
ENV HUSKY=0
RUN bun install --frozen-lockfile

FROM deps AS build
COPY apps/web apps/web
ENV NEXT_TELEMETRY_DISABLED=1
RUN bun run build

FROM oven/bun:1.4.2 AS runner
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000
COPY --from=build /app/package.json /app/bun.lock ./
COPY --from=build /app/node_modules node_modules
COPY --from=build --chown=bun:bun /app/apps/web apps/web
USER bun
EXPOSE 3000
# Applies pending migrations, then serves the production build.
CMD ["sh", "-c", "bun run db:migrate && bun run start"]
