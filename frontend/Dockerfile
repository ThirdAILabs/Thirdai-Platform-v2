FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY . /app
WORKDIR /app

FROM base AS prod-deps
RUN pnpm install --prod --frozen-lockfile
RUN pnpm add sharp @svgr/webpack  # Install the missing dependencies

FROM base
COPY --from=prod-deps /app/node_modules /app/node_modules

CMD [ "sh", "-c", "pnpm install --frozen-lockfile && pnpm run build && pnpm start" ]
