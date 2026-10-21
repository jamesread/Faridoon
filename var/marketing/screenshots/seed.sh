#!/usr/bin/env bash
# Seeds the throwaway backend with fictional users, settings, quotes and votes
# via its public API (Connect JSON). Usage: seed.sh <backend-base-url> <tmpdir>
#
# Every account uses the same throwaway password, which only exists inside the
# temporary stack. The first user registered on an empty database is the admin.
set -euo pipefail

base="${1:?backend base url}"
tmp="${2:?tmp dir}"
here="$(cd "$(dirname "$0")" && pwd)"
api="$base/faridoon.v1.FaridoonService"
password="marketing-demo"
jars="$tmp/cookies"

rm -rf "$jars"
mkdir -p "$jars"

# rpc <user> <method> <json>: calls a method with that user's session cookie.
rpc() {
	curl -fsS -b "$jars/$1" -c "$jars/$1" -X POST -H 'Content-Type: application/json' \
		-d "$3" "$api/$2" >/dev/null
}

jq -r '.[].username' "$here/seed/users.json" | while read -r user; do
	rpc "$user" Register "$(jq -nc --arg u "$user" --arg p "$password" \
		'{username: $u, password: $p, passwordConfirmation: $p}')"
done

admin="$(jq -r '.[0].username' "$here/seed/users.json")"

jq -c 'to_entries[] | {key, valueInt: .value}' "$here/seed/config.json" | while read -r cvar; do
	rpc "$admin" UpdateCvar "$cvar"
done

jq -c '.[]' "$here/seed/header-links.json" | while read -r link; do
	rpc "$admin" CreateHeaderLink "$link"
done

# Admin submissions are auto-approved; they get ids 1..n on the empty database.
jq -c '.[]' "$here/seed/quotes.json" | while read -r quote; do
	rpc "$admin" CreateQuote "$quote"
done

jq -c 'to_entries[] | .key as $id | .value | to_entries[] | [.key, $id, .value]' \
	"$here/seed/votes.json" | while read -r vote; do
	rpc "$(jq -r '.[0]' <<<"$vote")" VoteQuote "$(jq -c '{id: (.[1] | tonumber), delta: .[2]}' <<<"$vote")"
done

# Regular users' submissions wait in the approval queue.
jq -c '.[]' "$here/seed/pending-quotes.json" | while read -r quote; do
	rpc "$(jq -r '.by' <<<"$quote")" CreateQuote "$(jq -c '{content}' <<<"$quote")"
done

echo "Seeded $(jq length "$here/seed/quotes.json") quotes," \
	"$(jq length "$here/seed/pending-quotes.json") pending, as admin '$admin'."
