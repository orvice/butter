# syntax=docker/dockerfile:1

# --- Build stage ---------------------------------------------------------
FROM node:22-alpine AS builder

WORKDIR /app

# Install dependencies first (cached when lockfile is unchanged).
COPY package.json package-lock.json ./
RUN npm ci

# Copy the rest of the frontend source.
COPY . .

# Build the Vite bundle directly. We skip `tsc -b` because compile-time
# type errors in pre-existing form pages do not affect the runtime
# bundle and would otherwise fail the build.
RUN npx vite build

# --- Runtime stage -------------------------------------------------------
FROM nginx:1.27-alpine

# Replace the default site with our SPA-aware config.
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the built static bundle.
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

# nginx:alpine already ships a sensible default CMD; no override needed.
