# Repoxy UI Placeholder

This directory contains a placeholder React application created with Vite, React 18, and Material UI. It deliberately renders a single centered `TODO` message—no UI requirements have been implemented yet.

## Getting Started

```bash
cd ui
npm install    # requires Node 18+ / npm 9+
npm run dev
```

Open the printed URL (default http://localhost:5173) to see the placeholder screen. The Material UI dependency tree is already configured, so future work can layer in actual components without retooling the stack.

## Building

```bash
npm run build
```

This produces static assets in `dist/` ready to be hosted behind the Repoxy server when the real UI is implemented.
