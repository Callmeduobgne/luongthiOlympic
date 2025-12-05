# Chaincode as a Service (CCaaS) Dockerfile
# This runs chaincode as a standalone service, avoiding Docker-in-Docker issues

FROM node:18-alpine

# Install dependencies for native modules
RUN apk add --no-cache python3 make g++ git

WORKDIR /chaincode

# Copy package files first for better cache
COPY package*.json ./

# Install production dependencies only
RUN npm ci --omit=dev

# Copy built chaincode (dist folder)
COPY dist/ ./dist/
COPY package.json ./

# Set environment variables
ENV NODE_ENV=production
ENV CORE_CHAINCODE_ID_NAME=teaTraceCC:1.0
ENV CORE_PEER_TLS_ENABLED=true

# Expose chaincode port
EXPOSE 9999

# Start chaincode server
CMD ["node", "dist/index.js"]
