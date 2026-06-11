# krengki

CLI scaffolding tool. Currently supports creating Next.js projects via `npx create-next-app`.

## How it works

`krengki install` resolves the stable Next.js scaffold tag before running `npx`: it uses `create-next-app@lts` when npm publishes that tag, otherwise it uses the newest stable `next-*` dist-tag below `latest` (for example, `next-15-3` while `latest` is Next.js 16). After `npx create-next-app` finishes, it:

1. Copies all files from `cmd/dep/` into the project root (overwrites Next.js defaults on conflict)
2. Creates `.agents/skills` and `.agents/commands` directories
3. Symlinks `.claude/skills` → `../.agents/skills`, `.claude/commands` → `../.agents/commands`
4. Symlinks `.cursor/skills` → `../.agents/skills`, `.cursor/commands` → `../.agents/commands`

Add shared skills/commands to `cmd/dep/.agents/skills` and `cmd/dep/.agents/commands` — they get embedded into the binary at build time.

## Requirements

- Go 1.21+
- Node.js + npx

## Install

### curl (macOS / Linux)

```bash
curl -sSfL https://raw.githubusercontent.com/redhajuanda/krengki/main/install.sh | sh
```

To install a specific version:

```bash
VERSION=v0.1.0 curl -sSfL https://raw.githubusercontent.com/redhajuanda/krengki/main/install.sh | sh
```

### go install

```bash
go install github.com/redhajuanda/krengki@latest
```

Requires Go 1.21+. Binary lands in `$GOPATH/bin` (ensure it's in your `$PATH`).

### From source

```bash
git clone https://github.com/redhajuanda/krengki
cd krengki
make install
```

### Verify

```bash
krengki --version
```

## Usage

### `krengki install <project-name>`

Creates a new Next.js project using the stable/LTS `create-next-app` tag instead of npm `latest`.

```bash
krengki install my-app
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--typescript` | `true` | Use TypeScript |
| `--javascript` | `false` | Use JavaScript instead of TypeScript |
| `--tailwind` | `true` | Add Tailwind CSS |
| `--eslint` | `false` | Add ESLint |
| `--app` | `true` | Use App Router |
| `--src-dir` | `false` | Use `src/` directory structure |
| `--turbopack` | `false` | Enable Turbopack dev server |
| `--import-alias` | `""` | Custom import alias (e.g. `@/*`) |
| `--package-manager` | `pnpm` | Package manager to use (`npm`, `yarn`, `pnpm`, `bun`) |
| `--skip-install` | `false` | Skip package install |

#### Examples

Minimal (TypeScript only):
```bash
krengki install my-app
```

Full setup:
```bash
krengki install my-app --tailwind --eslint --src-dir --import-alias "@/*"
```

JavaScript project:
```bash
krengki install my-app --javascript --tailwind --eslint
```

Skip npm install:
```bash
krengki install my-app --tailwind --skip-install
```

### `krengki skills upgrade [version]`

Updates all dep-managed files (skills, commands, config) in an existing project.

```bash
# upgrade to latest release
krengki skills upgrade

# upgrade to a specific version
krengki skills upgrade v1.2.0

# upgrade a project in a different directory
krengki skills upgrade --dir ./my-app
krengki skills upgrade v1.2.0 --dir ./my-app
```

- Downloads dep files from GitHub release tarball
- Falls back to embedded dep if offline or no releases exist
- Recreates `.claude/skills`, `.claude/commands`, `.cursor/skills`, `.cursor/commands` symlinks
- Safe to run multiple times (idempotent)
