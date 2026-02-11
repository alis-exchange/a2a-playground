# a2a-playground Frontend

Vue 3 + Vuetify SPA for the A2A agent playground. Connects to the BFF via Connect-RPC.

## Stack

- **Vue 3** + **TypeScript**
- **Vuetify 3** – UI components
- **Vite** – build tooling
- **Pinia** – state management

## Development

```bash
pnpm dev
```

Serves at `http://localhost:3000` (or whatever port Vite assigns).

## Build

```bash
pnpm build
```

Output goes to `app/dist`, which is copied to `internal/bff/dist` for embedding in the CLI.
