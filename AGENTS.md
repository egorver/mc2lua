# Project Guidelines

## Coding Conventions

- Follow coding conventions defined in [docs/coding-conventions.md](docs/coding-conventions.md)

## Code Quality

- Run `go vet ./...` and `go test ./...` at the end of each work session

## Git Restrictions

- Git operations are strictly prohibited: no commits, no pushes, no branch operations, no reverts. Git is read-only.

## Testing

- Test the program with `go run ./cmd/mc2lua` (no arguments) or with specific bounds:
  `go run ./cmd/mc2lua -xmin 3 -xmax 126 -ymin -60 -ymax -5 -zmin 5 -zmax 123 -output output/output.lua`
