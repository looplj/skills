#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh --version vX.Y.Z [--remote origin] [--no-push]

What it does:
  1. Validates the worktree is clean.
  2. Uses the same version for all three modules.
  3. Publishes the root module tag first.
  4. Updates skillscmd to depend on github.com/looplj/skills@<version>, commits if needed, then tags and pushes skillscmd/<version>.
  5. Updates cmd/find-skills to depend on github.com/looplj/skills/skillscmd@<version>, commits if needed, then tags and pushes cmd/find-skills/<version>.

Examples:
  scripts/release.sh --version v0.0.2
  scripts/release.sh --version v0.0.2 --no-push
EOF
}

die() {
  echo "error: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

go_env_args=(
  "GOPRIVATE=github.com/looplj/*"
  "GONOPROXY=github.com/looplj/*"
  "GONOSUMDB=github.com/looplj/*"
  "GOPROXY=direct"
)

validate_version() {
  local value="$1"
  [[ "$value" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]] || die "invalid version: $value"
}

ensure_clean_tree() {
  local status
  status="$(git -C "$ROOT_DIR" status --porcelain)"
  [[ -z "$status" ]] || die "git worktree is not clean; commit or stash changes first"
}

ensure_tag_missing() {
  local tag="$1"
  if git -C "$ROOT_DIR" rev-parse -q --verify "refs/tags/$tag" >/dev/null 2>&1; then
    die "tag already exists: $tag"
  fi
}

maybe_commit_release_changes() {
  local message="$1"
  shift

  if git -C "$ROOT_DIR" diff --quiet -- "$@"; then
    return
  fi

  git -C "$ROOT_DIR" add "$@"
  git -C "$ROOT_DIR" commit -m "$message"
}

create_tag() {
  local tag="$1"
  local message="$2"
  git -C "$ROOT_DIR" tag -a "$tag" -m "$message"
}

push_commit_if_needed() {
  if [[ "$NO_PUSH" == "true" ]]; then
    return
  fi

  git -C "$ROOT_DIR" push "$REMOTE" HEAD
}

push_tag() {
  local tag="$1"
  if [[ "$NO_PUSH" == "true" ]]; then
    return
  fi

  git -C "$ROOT_DIR" push "$REMOTE" "refs/tags/$tag"
}

update_skillscmd_requirements() {
  (
    cd "$ROOT_DIR/skillscmd"
    go mod edit -require="github.com/looplj/skills@${VERSION}"
    env "${go_env_args[@]}" go mod tidy
  )
}

update_find_skills_requirements() {
  (
    cd "$ROOT_DIR/cmd/find-skills"
    go mod edit -require="github.com/looplj/skills/skillscmd@${VERSION}"
    env "${go_env_args[@]}" go mod tidy
  )
}

VERSION=""
REMOTE="origin"
NO_PUSH="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --remote)
      REMOTE="${2:-}"
      shift 2
      ;;
    --no-push)
      NO_PUSH="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

[[ -n "$VERSION" ]] || die "missing --version"

require_cmd git
require_cmd go

validate_version "$VERSION"

ensure_clean_tree
ensure_tag_missing "$VERSION"
ensure_tag_missing "skillscmd/$VERSION"
ensure_tag_missing "cmd/find-skills/$VERSION"

create_tag "$VERSION" "skills $VERSION"
push_commit_if_needed
push_tag "$VERSION"

update_skillscmd_requirements
maybe_commit_release_changes "release: skillscmd ${VERSION}" skillscmd/go.mod skillscmd/go.sum
push_commit_if_needed

create_tag "skillscmd/$VERSION" "skillscmd $VERSION"
push_tag "skillscmd/$VERSION"

update_find_skills_requirements
maybe_commit_release_changes "release: cmd/find-skills ${VERSION}" cmd/find-skills/go.mod cmd/find-skills/go.sum
push_commit_if_needed

create_tag "cmd/find-skills/$VERSION" "find-skills $VERSION"
push_tag "cmd/find-skills/$VERSION"

cat <<EOF
release prepared successfully
skills tag:         $VERSION
skillscmd tag:      skillscmd/$VERSION
find-skills tag:    cmd/find-skills/$VERSION
push enabled:       $([[ "$NO_PUSH" == "true" ]] && echo "no" || echo "yes")
remote:             $REMOTE
EOF
