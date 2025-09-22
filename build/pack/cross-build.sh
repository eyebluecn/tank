#!/bin/bash
#set -e

###########################################################################
#
#  Tank cross-compilation build script
#  Supports building for multiple platforms and architectures
#
###########################################################################

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -v, --version VERSION    Version name (default: tank-4.1.3)"
    echo "  -o, --os OS             Target OS: linux, windows, darwin, or 'all'"
    echo "  -a, --arch ARCH         Target architecture: amd64, arm64, 386, or 'all'"
    echo "  -p, --platform PLATFORM Specific platform (OS/ARCH format, e.g. linux/amd64)"
    echo "  --list-platforms        List all supported platforms"
    echo "  --clean                 Clean dist directory before building"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                              # Build for current platform"
    echo "  $0 -o linux -a amd64           # Build for Linux AMD64"
    echo "  $0 -p linux/arm64              # Build for Linux ARM64"
    echo "  $0 -o all -a amd64             # Build for all OS with AMD64"
    echo "  $0 -o all -a all               # Build for all supported platforms"
    echo "  $0 --clean -o all -a all       # Clean and build all platforms"
    exit 1
}

# List supported platforms
list_platforms() {
    echo "Supported platforms:"
    echo "  linux/amd64     - Linux 64-bit"
    echo "  linux/arm64     - Linux ARM64"
    echo "  linux/386       - Linux 32-bit"
    echo "  windows/amd64   - Windows 64-bit"
    echo "  windows/arm64   - Windows ARM64"
    echo "  windows/386     - Windows 32-bit"
    echo "  darwin/amd64    - macOS Intel"
    echo "  darwin/arm64    - macOS Apple Silicon"
    exit 0
}

# Log functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
VERSION_NAME="tank-4.1.3"
TARGET_OS=""
TARGET_ARCH=""
SPECIFIC_PLATFORM=""
BUILD_ALL_OS=false
BUILD_ALL_ARCH=false
CLEAN_DIST=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION_NAME="$2"
            shift 2
            ;;
        -o|--os)
            if [[ "$2" == "all" ]]; then
                BUILD_ALL_OS=true
            else
                TARGET_OS="$2"
            fi
            shift 2
            ;;
        -a|--arch)
            if [[ "$2" == "all" ]]; then
                BUILD_ALL_ARCH=true
            else
                TARGET_ARCH="$2"
            fi
            shift 2
            ;;
        -p|--platform)
            SPECIFIC_PLATFORM="$2"
            shift 2
            ;;
        --list-platforms)
            list_platforms
            ;;
        --clean)
            CLEAN_DIST=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Parse specific platform
if [[ -n "$SPECIFIC_PLATFORM" ]]; then
    if [[ "$SPECIFIC_PLATFORM" =~ ^([^/]+)/([^/]+)$ ]]; then
        TARGET_OS="${BASH_REMATCH[1]}"
        TARGET_ARCH="${BASH_REMATCH[2]}"
    else
        log_error "Invalid platform format. Use OS/ARCH format (e.g., linux/amd64)"
        exit 1
    fi
fi

# Set defaults if nothing specified
if [[ -z "$TARGET_OS" && "$BUILD_ALL_OS" == false ]]; then
    TARGET_OS=$(go env GOOS)
fi

if [[ -z "$TARGET_ARCH" && "$BUILD_ALL_ARCH" == false ]]; then
    TARGET_ARCH=$(go env GOARCH)
fi

# Define supported platforms as a simple list
SUPPORTED_PLATFORMS="linux/amd64 linux/arm64 linux/386 windows/amd64 windows/arm64 windows/386 darwin/amd64 darwin/arm64"

# Check if platform is supported
is_platform_supported() {
    local platform="$1"
    echo "$SUPPORTED_PLATFORMS" | grep -q "\b$platform\b"
}

# Validate platform
validate_platform() {
    local os=$1
    local arch=$2
    local platform="${os}/${arch}"

    if ! is_platform_supported "$platform"; then
        log_warn "Platform ${platform} may not be fully tested"
        return 1
    fi
    log_info "Platform ${platform} supported"
    return 0
}

# Setup directories
GOPROXY=https://goproxy.cn
PACK_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BUILD_DIR=$(dirname ${PACK_DIR})
PROJECT_DIR=$(dirname ${BUILD_DIR})
DIST_DIR=${PROJECT_DIR}/tmp/dist

log_info "=== Tank Cross-Compilation Build ==="
log_info "Version: ${VERSION_NAME}"
log_info "GOPROXY: ${GOPROXY}"
log_info "Project: ${PROJECT_DIR}"
log_info "Dist: ${DIST_DIR}"

# Clean dist directory if requested
if [[ "$CLEAN_DIST" == true ]]; then
    log_info "Cleaning dist directory..."
    rm -rf ${DIST_DIR}
fi

# Create dist directory
mkdir -p ${DIST_DIR}

# Build function for specific platform
build_platform() {
    local os=$1
    local arch=$2
    local platform="${os}/${arch}"

    log_info "Building ${platform}..."

    # Validate platform
    validate_platform "$os" "$arch"

    # Determine binary name and archive format
    local binary_name="tank"
    local archive_ext="tar.gz"
    if [[ "$os" == "windows" ]]; then
        binary_name="tank.exe"
        archive_ext="zip"
    fi

    # Setup paths
    local file_name="${VERSION_NAME}.${os}-${arch}.${archive_ext}"
    local component_dir="${DIST_DIR}/${VERSION_NAME}-${os}-${arch}"
    local dist_path="${DIST_DIR}/${file_name}"

    log_info "  Target: ${platform}"
    log_info "  Binary: ${binary_name}"
    log_info "  Output: ${file_name}"

    # Clean previous build
    if [[ -d "$component_dir" ]]; then
        rm -rf "$component_dir"
    fi
    mkdir -p "$component_dir"

    # Change to project directory
    cd ${PROJECT_DIR}

    # Cross-compile
    log_info "  Cross-compiling..."
    export GOOS="$os"
    export GOARCH="$arch"
    export GOPROXY="$GOPROXY"
    export CGO_ENABLED=0  # Disable CGO for cross-compilation

    log_info "  Starting go build..."  # 添加此行
    if ! go build -mod=readonly -ldflags="-s -w" -o "${binary_name}"; then
        log_error "Failed to build for ${platform}"
        return 1
    fi
    log_info "  go build completed"  # 添加此行

    # Copy binary
    log_info "  Copying binary..."
    cp "./${binary_name}" "${component_dir}/"

    # Copy configuration and static files
    log_info "  Copying build files..."
    if [[ -d "${BUILD_DIR}/conf" ]]; then
        cp -r "${BUILD_DIR}/conf" "${component_dir}/"
    fi
    if [[ -d "${BUILD_DIR}/html" ]]; then
        cp -r "${BUILD_DIR}/html" "${component_dir}/"
    fi

    # Create archive
    log_info "  Creating archive..."
    cd ${DIST_DIR}

    if [[ "$archive_ext" == "zip" ]]; then
        # Use zip for Windows
        if command -v zip >/dev/null 2>&1; then
            zip -r "${file_name}" "$(basename "$component_dir")" >/dev/null 2>&1
        else
            log_warn "zip command not found, using tar.gz instead"
            tar -zcf "${VERSION_NAME}.${os}-${arch}.tar.gz" "$(basename "$component_dir")" >/dev/null 2>&1
        fi
    else
        # Use tar.gz for Unix-like systems
        tar -zcf "${file_name}" "$(basename "$component_dir")" >/dev/null 2>&1
    fi

    if [[ $? -ne 0 ]]; then
        log_error "Failed to create archive for ${platform}"
        return 1
    fi

    # Cleanup
    cd ${PROJECT_DIR}
    rm -f "./${binary_name}"
    rm -rf "$component_dir"

    # Get file size
    local file_size=$(ls -lh "${DIST_DIR}/${file_name}" | awk '{print $5}')
    log_success "Built ${file_name} (${file_size})"

    return 0
}

# Build execution
build_count=0
success_count=0
failed_builds=()

if [[ "$BUILD_ALL_OS" == true && "$BUILD_ALL_ARCH" == true ]]; then
    log_info "Building for all supported platforms..."
    # Build all supported platforms explicitly
    for os in linux windows darwin; do
        for arch in amd64 arm64 386; do
            platform="${os}/${arch}"
            if is_platform_supported "$platform"; then
                ((build_count++))
                if build_platform "$os" "$arch"; then
                    ((success_count++))
                else
                    failed_builds+=("$platform")
                fi
            fi
        done
    done
elif [[ "$BUILD_ALL_OS" == true ]]; then
    log_info "Building for all OS with architecture: ${TARGET_ARCH}"
    for os in linux windows darwin; do
        platform="${os}/${TARGET_ARCH}"
        if is_platform_supported "$platform"; then
            ((build_count++))
            if build_platform "$os" "$TARGET_ARCH"; then
                ((success_count++))
            else
                failed_builds+=("$platform")
            fi
        fi
    done
elif [[ "$BUILD_ALL_ARCH" == true ]]; then
    log_info "Building for OS ${TARGET_OS} with all architectures"
    for arch in amd64 arm64 386; do
        platform="${TARGET_OS}/${arch}"
        if is_platform_supported "$platform"; then
            ((build_count++))
            if build_platform "$TARGET_OS" "$arch"; then
                ((success_count++))
            else
                failed_builds+=("$platform")
            fi
        fi
    done
else
    log_info "Building for ${TARGET_OS}/${TARGET_ARCH}"
    ((build_count++))
    if build_platform "$TARGET_OS" "$TARGET_ARCH"; then
        ((success_count++))
    else
        failed_builds+=("${TARGET_OS}/${TARGET_ARCH}")
    fi
fi

# Build summary
echo ""
log_info "=== Build Summary ==="
echo "Total builds: ${build_count}"
echo "Successful: ${success_count}"
echo "Failed: $((build_count - success_count))"

if [[ ${#failed_builds[@]} -gt 0 ]]; then
    echo ""
    log_warn "Failed builds:"
    for platform in "${failed_builds[@]}"; do
        echo "  - $platform"
    done
fi

if [[ $success_count -gt 0 ]]; then
    echo ""
    log_success "Generated files:"
    ls -lh ${DIST_DIR}/*.{tar.gz,zip} 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'

    echo ""
    log_info "Files located in: ${DIST_DIR}"
fi

echo ""
if [[ $((build_count - success_count)) -eq 0 ]]; then
    log_success "All builds completed successfully!"
    exit 0
else
    log_error "Some builds failed!"
    exit 1
fi