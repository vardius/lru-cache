# Contributing

Thanks for helping improve **lru-cache**. By contributing, you agree to abide by the [code of conduct](CODE_OF_CONDUCT.md).

## Issues and questions

Search the [issue tracker](https://github.com/vardius/lru-cache/issues) first; if nothing matches, open an issue.
Bugs, ideas and usage questions are all welcome. For a bug, include the calls you made, in order, and what you expected back.

## Development

You need Go 1.26 or newer.

```sh
git clone https://github.com/vardius/lru-cache.git
cd lru-cache
go test ./...
```

Before sending a pull request, run what CI runs:

```sh
gofmt -l .
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go test -race -shuffle=on ./...
go test -run Allocations -count 1 ./...
```

### Layout

- `cache.go` holds `Cache` and its methods; `doc.go` is the package documentation.
- `TestAgainstModel` checks the cache against a slow, obviously correct one with random operations. A change to the
  index or the list should keep it passing; add a case to the table tests for any bug it finds.
- `examples/` holds runnable programs, each with a test. The website shows them, so keep them short.
- `website/` is the documentation site.

### Performance

`TestAllocations` fails if a call allocates more than it should; if your change moves a number, explain why in the pull request.
To compare before and after:

```sh
go test -run '^$' -bench . -count 10 > new.txt
```

and compare with the same run on `master`, for example with `golang.org/x/perf/cmd/benchstat`.

### Website

```sh
cd website
npm ci
npm run dev
```

## Pull requests

- For a large change, open an issue first so we can agree on the approach.
- Branch from `master`, keep the change focused, and add a test that fails without it.
