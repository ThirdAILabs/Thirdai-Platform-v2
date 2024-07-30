FROM node:21-bullseye-slim

WORKDIR /app

# Install `serve`
RUN npm install -g serve

# Copy build artifacts from local build directory
COPY ./build .
RUN mkdir -p ./search
RUN mv ./static ./search

# Start the application on port 5000
CMD ["serve", "-s", ".", "-l", "5000"]