# Butter Frontend

React + TypeScript + Vite dashboard for Butter.

## Local development

Start the backend first from the repository root:

```bash
cp .env.example .env
# If config/butter.yaml does not exist, set BUTTERFLY_CONFIG_FILE_PATH=./config.yaml
export $(grep -v '^#' .env | xargs)
go run ./cmd/butter
```

Verify the backend:

```bash
curl http://127.0.0.1:8080/ping
```

Then start the frontend:

```bash
cd front
cp .env.example .env.local
npm install
npm run dev
```

Open <http://localhost:5173>.

## API base URL and local proxy

The frontend supports two local API modes.

### Recommended: Vite dev proxy

For local development, keep `VITE_API_BASE_URL` empty and set the proxy target:

```env
VITE_API_BASE_URL=
VITE_DEV_PROXY_TARGET=http://localhost:8080
```

With this setup, browser requests use relative paths and go through the Vite dev server:

```text
Browser -> http://localhost:5173/api/... -> Vite proxy -> http://localhost:8080/api/...
Browser -> http://localhost:5173/ping -> Vite proxy -> http://localhost:8080/ping
```

The proxy target is read in `vite.config.ts` via `loadEnv()`. If `VITE_DEV_PROXY_TARGET` is not set, it defaults to `http://localhost:8080`.

### Direct API calls

If `VITE_API_BASE_URL` is set, browser requests call that origin directly instead of using the Vite proxy:

```env
VITE_API_BASE_URL=http://localhost:8080
```

In that mode, ConnectRPC requests are sent to:

```text
http://localhost:8080/api/...
```

Use this mode only when you intentionally want direct cross-origin requests and the backend allows them.

## Scripts

```bash
npm run dev      # Start Vite dev server
npm run build    # Type-check and build production assets
npm run lint     # Run ESLint
npm run preview  # Preview production build locally
```
