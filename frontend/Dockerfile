FROM node:24.13.1-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . ./
RUN npm run build

FROM node:24.13.1-alpine AS runtime

ENV NODE_ENV=production
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --omit=dev --ignore-scripts
COPY --from=build /app/.next ./.next

USER node
EXPOSE 3000
HEALTHCHECK --interval=15s --timeout=3s --start-period=15s --retries=5 CMD wget -q -O /dev/null http://127.0.0.1:3000/pt-PT || exit 1

CMD ["npm", "run", "start", "--", "--hostname", "0.0.0.0", "--port", "3000"]
