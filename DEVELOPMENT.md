# Development

## Required Tools for Development
The following tools are required for developing of this repo:

- [Golang](https://go.dev/dl/)
- [Nushell](https://www.nushell.sh/book/installation.html)
- [Just](https://just.systems/man/en/installation.html)
- [Golangci-lint](https://golangci-lint.run/docs/welcome/install/local/)

It can be helpful to use [mise](https://mise.jdx.dev/) for managing versioned dependencies that tend tro differ across repos like golangci-lint.

## Using Just

The following recipes are available - and you can get an up to date list by running `just` or `just help`:

```shell
> just

Available recipes:
    build                 # Build for Linux (amd64).
    build-all             # Build every supported platform.
    build-linux           # Build for Linux (amd64).
    build-mac             # Build for macOS (arm64).
    build-windows         # Build for Windows (amd64).
    ci                    # A quick and convenient wrapper to run a bunch of common commands before CI.
    clean                 # Remove the build output.
    cover                 # Open the coverage profile from `just test` in a browser.
    data-extract game_dir # Extract the source xml from a KCD2 install, then rebuild data/dice.json. Must provide the path to your game data.
    data-regenerate       # Rebuild data/dice.json from the xml files already in ./data.
    fix                   # Apply fixes for outdated APIs.
    fmt                   # Format the code.
    help                  # Prints all available recipes.
    lint                  # Lint the code, applying fixes where possible.
    test                  # Run the test suite with the race detector and per-package coverage.
    vet                   # Identify insecure code.
```

## Updating Tool Versions
If you update the version of a tool in use (such as `golangci-lint`) then be sure to update the [mise.toml](mise.toml) as our CI will leverage that.

If you are using `mise` - then you can use the following commands:
```shell
# explicit version change
mise use golangci-lint@2.14.0

# bump everything to latest within what your pin allows
mise upgrade --bump

# check for outdated deps
mise outdated             # what's behind, and what's available 
mise latest golangci-lint # newest version that exists
```

## Cutting a Release

Releases are driven from the GitHub release itself - publishing one creates the tag and kicks off
[release.yml](.github/workflows/release.yml), which builds the Linux, macOS and Windows binaries, attaches them to the
release, and then tells the [pkgs](https://github.com/treyburn/pkgs) repo to redeploy
[go.treyburn.dev](https://go.treyburn.dev).

1. Merge everything you want in the release to `main`.
2. Draft a new release on GitHub, targeting `main`, with a `vX.Y.Z` tag.
3. Write the notes and publish it.
4. Watch the `Release` workflow - it attaches the three binaries and triggers the vanity deploy.

The version reported by `farkle --version` is not set by hand - Go stamps it from the tag at build time, so the tag is
the only place the version lives. A release built from `v0.1.1` reports `farkle v0.1.1`.

> [!NOTE]
> Leave immutable releases disabled. When enabled, GitHub seals a release the moment it is published and the workflow
> can no longer attach the binaries - which is how `v0.1.0` shipped with no assets at all.

## Updating the Source Data
The dice tables are generated from the game's own data files and embedded into the binary at build time, so refreshing them is a source change rather than a build step - run it when a game patch lands and commit the result.

```shell
# extract the xml from a KCD2 install and regenerate data/dice.json
just data-extract ~/.steam/steam/steamapps/common/KingdomComeDeliverance2

# regenerate data/dice.json from the xml data already in ./data
just data-regenerate
```

See [data/README.md](data/README.md) for where the game keeps those files, what the pipeline does with them, and the exploration notes behind it.
