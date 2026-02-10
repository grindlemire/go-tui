# go-tui docs site

Minimal Docusaurus placeholder configured for Vercel.

## Local development

```bash
cd docs/site
npm install
npm run start
```

## Build

```bash
cd docs/site
npm run build
```

## Vercel setup

1. Import this repository into Vercel.
2. Set **Root Directory** to `docs/site`.
3. Keep defaults from `vercel.json`.
4. Deploy.

Optional updates before first production deploy:

- Set `url` in `docusaurus.config.js` to your real domain.
- If using a custom domain path prefix, set `baseUrl` accordingly.
