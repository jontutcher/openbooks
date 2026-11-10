#!/bin/sh
# Smoke test for the image .gitlab-ci.yml has just built. Nothing is pushed
# that fails this.
#
#   sh .gitlab/smoke-test.sh <image>
#
# The container is started the way the cluster starts it: nonroot, read-only
# root filesystem, and only /tmp and the download directory writable. The HTTP
# probe runs from a sibling container on a user-defined network rather than
# through a published port, because under docker-in-docker the published port
# lands on the dind service, not on the job container's localhost.
set -eu

IMAGE="${1:?usage: smoke-test.sh <image>}"
NAME="openbooks-smoke-$$"
NET="$NAME-net"

cleanup() {
    docker logs "$NAME" 2>&1 | sed 's/^/    [container] /' || true
    docker rm -f "$NAME" >/dev/null 2>&1 || true
    docker network rm "$NET" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker network create "$NET" >/dev/null

# --name is the IRC nick, which the CLI requires; the server serves HTTP
# without connecting to IRC, which only happens once a browser attaches.
docker run -d --name "$NAME" --network "$NET" \
    --read-only --tmpfs /tmp --tmpfs /books \
    "$IMAGE" --name "$NAME" >/dev/null

fetch() {
    docker run --rm --network "$NET" alpine:3 sh -c \
        "for i in \$(seq 1 60); do wget -q -O- 'http://$NAME:80$1' && exit 0; sleep 1; done; exit 1"
}

echo "==> index.html is served"
INDEX=$(fetch /)
echo "$INDEX" | grep -q '<title>OpenBooks</title>'

# The client is compiled into the binary with //go:embed, so a build that
# forgot the web stage still starts and still answers on /: only the hashed
# asset actually proves the bundle made it in.
echo "==> the embedded client bundle is served"
# vite emits the reference relative to the page ("./assets/index-<hash>.js"),
# so the leading "./" is stripped before the path is requested.
BUNDLE=$(echo "$INDEX" | sed -n 's|.*src="[./]*\(assets/index-[^"]*\.js\)".*|\1|p' | head -1)
[ -n "$BUNDLE" ] || { echo "no hashed bundle referenced by index.html"; exit 1; }
BYTES=$(fetch "/$BUNDLE" | wc -c)
[ "$BYTES" -gt 100000 ] || { echo "$BUNDLE is only $BYTES bytes"; exit 1; }
echo "    /$BUNDLE ($BYTES bytes)"

echo "==> it runs as nonroot"
USER_SPEC=$(docker inspect "$NAME" --format '{{.Config.User}}')
[ "$USER_SPEC" = "65532" ] || { echo "expected uid 65532, got '$USER_SPEC'"; exit 1; }

echo "==> downloads are written to /books"
docker logs "$NAME" 2>&1 | grep -q 'Download Directory: /books'

echo "==> it is still running"
[ "$(docker inspect "$NAME" --format '{{.State.Running}}')" = "true" ]

echo "OK: $IMAGE"
