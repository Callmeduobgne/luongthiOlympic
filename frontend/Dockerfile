# Stage 1: Build
FROM node:20-alpine AS builder

# Build arguments for version injection
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

# Build metadata
LABEL maintainer="IBN Network Team"
LABEL description="IBN Frontend - Blockchain traceability UI"
LABEL version="${VERSION}"
LABEL build-date="${BUILD_DATE}"
LABEL git-commit="${GIT_COMMIT}"

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build application
RUN npm run build

# Stage 2: Production
FROM nginx:alpine

# Build arguments for labels
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

# Build metadata
LABEL maintainer="IBN Network Team"
LABEL description="IBN Frontend - Blockchain traceability UI"
LABEL version="${VERSION}"
LABEL build-date="${BUILD_DATE}"
LABEL git-commit="${GIT_COMMIT}"

# Install wget for healthcheck
RUN apk add --no-cache wget

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

