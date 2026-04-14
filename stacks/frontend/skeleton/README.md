# Project Accelerator — React Router 7 Frontend Skeleton

This is an executable React Router 7 starter that serves as the foundation for any new frontend project. It is extracted from a production multi-tenant application, stripped of all domain-specific logic, and left as a clean runway for building on top of.

You can copy this folder, run `pnpm install && pnpm dev`, and immediately see a working application at `http://localhost:3100`.

## What is included

The skeleton ships with everything a production frontend needs before the first feature is written:

- **React Router 7 in SSR mode** with streaming server rendering, bot detection, and proper hydration
- **Tailwind CSS v4** using the new CSS-based configuration with `@theme` directives
- **A theme system** that supports light, dark, and system preferences — applied before hydration to prevent flash of wrong theme
- **A component library** (Button, Card, Input, Label, Badge, Separator, Spinner) built with CVA + Radix Slot, following shadcn/ui conventions
- **A sidebar layout** with navigation sections, active state styling, and a theme toggle
- **TypeScript in strict mode** with path aliases (`~/`) for clean imports
- **ESLint v9 flat config** with import ordering, accessibility, and consistent type imports
- **Prettier** with Tailwind CSS class sorting
- **Vitest + Testing Library** configured and ready
- **Docker support** with a multi-stage build

## Running it

```bash
# Copy the skeleton to your project location
cp -r stacks/frontend/skeleton/ ~/my-new-project
cd ~/my-new-project

# Install dependencies
pnpm install

# Start development server (port 3100)
pnpm dev
```

## Available scripts

| Script | What it does |
|--------|-------------|
| `pnpm dev` | Start development server with HMR on port 3100 |
| `pnpm build` | Production build |
| `pnpm start` | Serve the production build |
| `pnpm typecheck` | Generate route types and run TypeScript compiler |
| `pnpm lint` | Run ESLint |
| `pnpm lint:fix` | Run ESLint with auto-fix |
| `pnpm format` | Format all files with Prettier |
| `pnpm test` | Run tests with Vitest |
| `pnpm test:coverage` | Run tests with v8 coverage |

## Architecture

The project follows a domain-oriented structure where each feature area lives in its own folder under `app/pages/`:

```
app/
  components/          # Shared components
    ui/                # Primitive UI components (Button, Card, etc.)
    layouts/           # Layout shells (sidebar, theme toggle)
  lib/                 # Shared utilities (cn, theme, constants)
  types/               # Shared TypeScript types
  pages/               # Domain folders
    app/               # The "app" domain (main authenticated shell)
      routes/          # Route components for this domain
  styles/              # Global CSS (Tailwind entry point)
  root.tsx             # Root layout (HTML shell, meta, theme bootstrap)
  routes.ts            # Route configuration (maps URLs to page components)
  entry.client.tsx     # Client hydration entry
  entry.server.tsx     # SSR streaming entry
```

When adding a new domain — say, "settings" — you would create `app/pages/settings/routes/` with its route components, then register them in `app/routes.ts`. Domain-specific components, hooks, and types live alongside the routes in that domain folder.

## Adding new features

1. Create a domain folder: `app/pages/my-feature/routes/`
2. Add route components inside it
3. Register routes in `app/routes.ts` using `route()` or `layout()`
4. Add sidebar navigation items in `_app-layout.tsx`

For shared UI, add components to `app/components/ui/` and export them from the barrel `index.ts`.

## Theme system

The theme system supports three modes: light, dark, and system (follows OS preference). It works by:

1. An inline script in `<head>` reads the stored preference from localStorage and applies the `dark` class to `<html>` before any rendering happens
2. The `ThemeToggle` component cycles through system, light, and dark on click
3. Tailwind's `@custom-variant dark` directive maps dark mode to the `.dark` class

This prevents any flash of wrong theme during SSR hydration because the class is applied synchronously before the browser paints.

## Testing

Tests use Vitest with jsdom and Testing Library. Place test files alongside the code they test using the `.test.tsx` or `.test.ts` extension:

```bash
# Run all tests
pnpm test

# Run with coverage
pnpm test:coverage

# Run in watch mode during development
npx vitest
```

## Environment variables

Copy `env.example` to `.env` and fill in values:

- `APP_ACCESS_PASSWORD` — optional access gate password
- `SESSION_COOKIE_SECRET` — secret for signing session cookies
- `API_URL` — backend API base URL

## Docker

Build and run with Docker:

```bash
docker build -t my-app .
docker run -p 3100:3100 my-app
```

The Dockerfile uses a multi-stage build: install dependencies, build the app, then serve from a slim production image.
