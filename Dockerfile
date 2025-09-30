# Multi-stage build for Aseprite CLI
FROM ubuntu:22.04 AS builder

# Prevent interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies
RUN apt-get update && apt-get install -y \
    cmake \
    ninja-build \
    g++ \
    libx11-dev \
    libxcursor-dev \
    libxi-dev \
    libgl1-mesa-dev \
    libfontconfig1-dev \
    libxrandr-dev \
    curl \
    unzip \
    git \
    && rm -rf /var/lib/apt/lists/*

# Download and extract Skia
WORKDIR /build
RUN curl -L https://github.com/aseprite/skia/releases/download/m124-08a5439a6b/Skia-Linux-Release-x64.zip -o skia.zip \
    && unzip skia.zip -d skia \
    && rm skia.zip

# Clone Aseprite
RUN git clone --recursive --branch v1.3.15.3 https://github.com/aseprite/aseprite.git

# Build Aseprite CLI (headless, no UI)
WORKDIR /build/aseprite/build
RUN cmake -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DENABLE_UI=OFF \
    -DENABLE_TESTS=OFF \
    -DENABLE_SCRIPTING=ON \
    -DLAF_BACKEND=none \
    -DSKIA_DIR=/build/skia \
    -DSKIA_LIBRARY_DIR=/build/skia/out/Release-x64 \
    .. \
    && ninja aseprite

# Final stage - minimal runtime image
FROM ubuntu:22.04

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    libx11-6 \
    libxcursor1 \
    libxi6 \
    libgl1 \
    libfontconfig1 \
    libxrandr2 \
    && rm -rf /var/lib/apt/lists/*

# Copy Aseprite binary, data files, and required libraries from builder
COPY --from=builder /build/aseprite/build/bin/aseprite /usr/local/bin/aseprite
COPY --from=builder /build/aseprite/build/bin/data /usr/local/share/aseprite/data
COPY --from=builder /build/skia/out/Release-x64/*.so* /usr/local/lib/

# Set Aseprite data directory
ENV ASEPRITE_USER_FOLDER=/tmp/aseprite

# Update library cache
RUN ldconfig

# Verify installation works in batch mode
RUN aseprite --batch --list-layers 2>&1 | grep -q "No document to execute the script" || aseprite --version

# Set aseprite as default command
ENTRYPOINT ["/usr/local/bin/aseprite"]
CMD ["--version"]