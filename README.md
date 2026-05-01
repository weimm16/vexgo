# vexgo (refactoring ...)

## Development Steps

### Prerequisites
- Go 1.25.7+ (with CGO enabled, required by `mattn/go-sqlite3`)
- Node.js + pnpm (for frontend development)

### 1. Clone the repository
```bash
git clone -b refactor https://github.com/vexgo-org/vexgo.git
cd vexgo
```

### 2. Frontend Development (Optional)
If you need to modify the admin frontend:
```bash
cd admin-frontend
pnpm install
pnpm build  # base path /admin/
rm -rf internal/admin/dist
cp -r admin-frontend/dist internal/admin/dist
```

### 3. Backend Development
- Install dependencies:
  ```bash
  go mod download
  ```
- Run locally:
  ```bash
  go run ./cmd/server/main.go
  # or
  go build -o vexgo ./cmd/server/
  ./vexgo
  ```
