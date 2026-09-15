#!/bin/sh
# Signs and notarizes a darwin binary with credentials from the login keychain.
# Called from the goreleaser build post hook: macos-sign.sh <os> <path> <snapshot>
#
# One-time setup:
#   xcrun notarytool store-credentials ink-notary \
#     --apple-id <apple-id> --team-id F34DLG5JYF --password <app-specific-password>
set -eu

os=$1
path=$2
snapshot=$3

[ "$os" = darwin ] || exit 0
[ "$snapshot" = true ] && exit 0

identity=${MACOS_SIGN_IDENTITY:-Developer ID Application}
profile=${MACOS_NOTARY_PROFILE:-ink-notary}

codesign --force --timestamp --options runtime --sign "$identity" "$path"
codesign --verify --strict "$path"

# notarytool only accepts zip, pkg or dmg uploads.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
ditto -c -k --keepParent "$path" "$tmp/ink.zip"

result=$(xcrun notarytool submit "$tmp/ink.zip" --keychain-profile "$profile" --wait --output-format json)
echo "$result"
echo "$result" | grep -q '"status":"Accepted"'
