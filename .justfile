set dotenv-load

# Build CLI binary by using Profile-Guided Optimization (PGO).
build:
    @touch cpuprofile.pprof
    @go build \
    -ldflags "-s -w" \
    -pgo cpuprofile.pprof \
    -o ./bin/$(basename $(pwd)) ./cmd/cli/main.go

# Profile the binary and generate a profile-guided optimization (PGO) profile.
profile:
    @rm -f cpuprofile.pprof
    @for pkg in \
    cmd/cli \
    internal/adapters/inbound \
    internal/adapters/outbound; \
    do \
        suffix_with_underscore=$(echo "$pkg" | tr '/\\' '__'); \
        out="cpuprofile-$suffix_with_underscore.pprof"; \
        go test ./$pkg/... \
          -run=^$ \
          -bench=. \
          -benchtime=10s \
          -cpuprofile="$out" \
          -pgo=off \
          2>&1 \
          >/dev/null || exit $?; \
    done

    @go tool pprof -proto cpuprofile-*.pprof > cpuprofile-merged.pprof
    @cp cpuprofile-merged.pprof cpuprofile.pprof
    @go tool pprof -svg cpuprofile.pprof > cpuprofile.svg
    @rm -f cpuprofile-*.pprof *.test

# Run the CLI.
run: build
    @./bin/$(basename $(pwd))

# Run the server.
serve:
    @go run ./cmd/server/main.go

# Test the Go sources (Units).
test:
    @go test -v -coverprofile=./coverage.pprof ./internal/...
    @echo "test coverage: $(go tool cover -func=coverage.pprof | grep total | awk '{print $3}')"
