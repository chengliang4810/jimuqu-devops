#!/usr/bin/env bash

set -euo pipefail

APP_NAME="jimuqu-devops"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
ARCHIVE_DIR="$BUILD_DIR/archives"
STAGE_DIR="$BUILD_DIR/staging"
FRONTEND_DIR="$ROOT_DIR/web-next"
EMBED_DIR="$ROOT_DIR/internal/httpapi/webdist"

RAW_VERSION="${1:-${GITHUB_REF_NAME:-$(git -C "$ROOT_DIR" describe --tags --abbrev=0 2>/dev/null || git -C "$ROOT_DIR" rev-parse --short HEAD)}}"
VERSION="${RAW_VERSION#v}"
COMMIT="$(git -C "$ROOT_DIR" rev-parse --short HEAD)"
BUILD_TIME="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

TARGETS=(
  "linux amd64 x86_64"
  "linux 386 x86"
  "linux arm64 arm64"
  "linux arm armv7"
  "darwin amd64 x86_64"
  "darwin arm64 arm64"
  "windows amd64 x86_64"
  "windows 386 x86"
  "windows arm64 arm64"
)

archive_stage() {
  local source_dir="$1"
  local archive_path="$2"

  if command -v zip >/dev/null 2>&1; then
    (
      cd "$source_dir"
      zip -qr "$archive_path" .
    )
    return 0
  fi

  if command -v powershell >/dev/null 2>&1; then
    local windows_source
    local windows_archive
    windows_source="$(cd "$source_dir" && pwd -W 2>/dev/null || pwd)"
    windows_archive="$(cd "$(dirname "$archive_path")" && pwd -W 2>/dev/null || pwd)\\$(basename "$archive_path")"
    powershell -NoProfile -Command "Compress-Archive -Path '$windows_source\\*' -DestinationPath '$windows_archive' -Force" >/dev/null
    return 0
  fi

  echo "No zip implementation found. Install zip or run in PowerShell-enabled environment." >&2
  return 1
}

rm -rf "$BUILD_DIR"
mkdir -p "$ARCHIVE_DIR" "$STAGE_DIR"

cleanup_embed_dir() {
  find "$EMBED_DIR" -mindepth 1 ! -name '.gitignore' ! -name '.keep' -exec rm -rf {} +
}

sync_embed_assets() {
  cleanup_embed_dir
  cp -R "$FRONTEND_DIR/out/." "$EMBED_DIR/"
}

pushd "$FRONTEND_DIR" >/dev/null
pnpm install --frozen-lockfile
NEXT_PUBLIC_APP_VERSION="$VERSION" pnpm build
popd >/dev/null

sync_embed_assets
trap cleanup_embed_dir EXIT

for target in "${TARGETS[@]}"; do
  read -r GOOS GOARCH ARCH_LABEL <<<"$target"

  PACKAGE_NAME="${APP_NAME}-${GOOS}-${ARCH_LABEL}"
  STAGE_PATH="$STAGE_DIR/$PACKAGE_NAME"
  mkdir -p "$STAGE_PATH"

  BINARY_NAME="server"
  if [[ "$GOOS" == "windows" ]]; then
    BINARY_NAME="server.exe"
  fi

  pushd "$ROOT_DIR" >/dev/null
  GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X 'devops-pipeline/internal/version.Version=$VERSION' -X 'devops-pipeline/internal/version.Commit=$COMMIT' -X 'devops-pipeline/internal/version.BuildTime=$BUILD_TIME'" \
    -o "$STAGE_PATH/$BINARY_NAME" \
    ./cmd/server
  popd >/dev/null

  cp "$ROOT_DIR/README.md" "$STAGE_PATH/"

  archive_stage "$STAGE_PATH" "$ARCHIVE_DIR/$PACKAGE_NAME.zip"
done
