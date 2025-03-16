FROM node:20-alpine AS base

WORKDIR /app

COPY package*.json ./

# Clean npm cache for smaller image
RUN npm ci && npm cache clean --force

COPY . .

EXPOSE 4200

CMD ["npm", "run", "start", "--", "--host", "0.0.0.0", "--poll", "500"]