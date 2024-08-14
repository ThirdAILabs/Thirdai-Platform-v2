FROM node:21-bullseye-slim

WORKDIR /app

# Copy build artifacts from local build directory
COPY ./.next .next
COPY ./package.json package.json
COPY ./package-lock.json package-lock.json
COPY ./public public

RUN npm install --force

# Start the application on port 3000
CMD ["npm", "start"]