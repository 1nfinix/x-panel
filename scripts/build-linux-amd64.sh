#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version="$(cut -d '|' -f 2 "$repo_root/service/assets/version" | tr -d '\r\n')"
if [[ -z "$version" ]]; then
  echo "Missing version in service/assets/version" >&2
  exit 1
fi

package_name="x-panel_v${version}_linux_amd64"
release_dir="$repo_root/release"
mkdir -p "$release_dir"
staging="$(mktemp -d "$release_dir/.build-linux-amd64.XXXXXX")"
trap 'rm -rf "$staging"' EXIT

docker buildx build --platform linux/amd64 \
  -f "$repo_root/Dockerfile.build" \
  --output "type=local,dest=$staging/$package_name" \
  "$repo_root"

archive="$release_dir/$package_name.tar.gz"
# macOS bsdtar would otherwise embed xattrs as ._* AppleDouble files that GNU tar extracts as junk.
COPYFILE_DISABLE=1 tar --no-xattrs -C "$staging" -czf "$archive" "$package_name"
shasum -a 256 "$archive"
