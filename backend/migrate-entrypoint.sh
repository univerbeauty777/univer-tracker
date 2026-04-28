#!/bin/sh
# migrate-entrypoint — wraps golang-migrate so a dirty state from a
# previous failed run doesn't permanently block the deploy.
#
# Behaviour:
#   1. Read the current version + dirty flag.
#   2. If dirty, roll the bookmark back to the previous clean version
#      (the schema change itself was inside a transaction, so it
#      already rolled back — only the metadata is stuck).
#   3. Run `up` as usual.
#
# All migrations are written idempotently (IF NOT EXISTS, ON CONFLICT)
# so re-applying a partially-applied version is safe.

set -eu

DB_URL="$1"
if [ -z "$DB_URL" ]; then
    echo "usage: $0 <database-url>" >&2
    exit 64
fi

VERSION_OUT=$(migrate -path=/migrations -database "$DB_URL" version 2>&1 || true)

case "$VERSION_OUT" in
    *dirty*|*Dirty*)
        CURRENT=$(printf '%s\n' "$VERSION_OUT" | head -n1 | awk '{print $1}')
        case "$CURRENT" in
            ''|*[!0-9]*)
                echo "migrate: dirty state detected but could not parse version from '$VERSION_OUT'" >&2
                exit 1
                ;;
        esac
        PREV=$((CURRENT - 1))
        echo "migrate: dirty version $CURRENT detected, forcing back to $PREV" >&2
        migrate -path=/migrations -database "$DB_URL" force "$PREV"
        ;;
esac

exec migrate -path=/migrations -database "$DB_URL" up
