---
sidebar_position: 2
---

# Theme Showcase

This page intentionally mixes common doc elements to validate style coverage.

## Typography and emphasis

Normal paragraph text should stay easy to read. Inline code like `AppOptions` should be distinct. Highlighted text like <mark>critical configuration</mark> should feel visible but not noisy.

> Good docs style: high contrast, clear spacing, obvious hierarchy.

## Status callouts

:::tip Working well
Dark mode glow is limited to interactive elements so body copy remains clear.
:::

:::warning Review before launch
Set the production URL in `docusaurus.config.js` before first public deploy.
:::

:::danger Common pitfall
Avoid neon yellow for light-mode link text. It fails contrast quickly.
:::

## Code blocks

```bash
bun run build
bun run serve
```

```json
{
  "framework": "docusaurus-2",
  "buildCommand": "npm run build",
  "outputDirectory": "build"
}
```

## Checklist

- Verify link color contrast in light mode.
- Verify glow intensity in dark mode.
- Verify heading hierarchy on mobile width.
- Verify code block readability in both themes.

## Sample API table

| Item | Type | Description |
| --- | --- | --- |
| `App` | Struct | Main runtime container |
| `Element` | Interface | Renderable/composable unit |
| `OnMount` | Hook | Lifecycle callback at mount time |
| `Dispatch` | Function | Event delivery entry point |

## Keyboard sample

Use <kbd>Shift</kbd> + <kbd>D</kbd> to toggle dark mode while reviewing contrast.
