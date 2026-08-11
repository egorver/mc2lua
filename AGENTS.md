# Project Guidelines

## Coding Conventions

- Follow coding conventions defined in [docs/coding-conventions.md](docs/coding-conventions.md)

## Code Quality

- Run `go vet ./...`, `go test ./...`, and `golangci-lint run ./...` at the end of each work session

## Git Restrictions

- Git operations are strictly prohibited: no commits, no pushes, no branch operations, no reverts. Git is read-only.

## Testing

- Test the program with `go run ./cmd/mc2lua` (no arguments) or with specific bounds:
  `go run ./cmd/mc2lua -xmin -2201 -xmax -2125 -ymin 60 -ymax 110 -zmin 2669 -zmax 2722 -output output/output.lua -parts-template output/template.yaml`
