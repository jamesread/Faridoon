# Marketing screenshots

The images in this directory are used in the project README and docs. They are
captured from a throwaway Faridoon stack seeded with fictional data, so every
run produces the same pictures (only the dates move with the current day).

## Requirements

- `podman` (or `docker`: `make CONTAINER=docker`)
- `sql-migrate`, `jq`, `curl`
- `repo-helper` with Chrome/Chromium and a matching `chromedriver`
- The normal build toolchain (`go`, `npm`, `buf`), because `make` builds the app first

## Usage

```sh
make            # build, start, seed, capture, write the manifest, publish, tear down
make up seed    # just start a seeded stack to browse at http://127.0.0.1:18480/
make down       # stop it again
```

Sign in as `morgan` / `marketing-demo` to browse as the admin. That account
exists only inside the temporary stack.

## How it works

- `make up` starts MySQL in a temporary container (`faridoon-mkt-mysql`, port
  13306, no volumes), applies the migrations and runs the built service on
  127.0.0.1:18480. The service serves the built frontend itself.
- `seed.sh` registers fictional users and adds settings, a header link, quotes,
  votes and pending submissions through the public API, using the files in
  `seed/`. `seed/quote-history.sql` then spreads the quotes' creation dates over
  the previous days, which the API can't do.
- `screenshots.ini` lists the captures. Its `title`, `description` and `theme`
  keys are turned into `screenshot-manifest.yaml` by `generate-manifest.py`.
  Don't edit the manifest by hand.
- Generated config, logs and session files go in `tmp/`, which is gitignored.
