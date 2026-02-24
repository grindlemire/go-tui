# Changelog

## 1.0.0 (2026-02-24)


### Features

* add ClickBinding type for ref-based mouse handling ([a18bf5b](https://github.com/grindlemire/go-tui/commit/a18bf5b0b90f47e4e807845db9e9c5d41031fb44))
* add generic NewChannelWatcher helper ([52f3f86](https://github.com/grindlemire/go-tui/commit/52f3f86082ef2285a04f5b684ada9ceabf92edc7))
* add HandleClicks helper for automatic ref hit testing ([ba91e9d](https://github.com/grindlemire/go-tui/commit/ba91e9d9d6b79c06cb16eb6198ad635450bfa5fa))
* add tree walk to collect watchers from WatcherProvider components ([2abe659](https://github.com/grindlemire/go-tui/commit/2abe659c861332e3f8cc3772c2f1e4b3d88244f6))
* add WatcherProvider interface for component-level watchers ([4ceac54](https://github.com/grindlemire/go-tui/commit/4ceac54f40f893b1531a4f83adc69a69a409974f))
* **ai-chat:** add ChatApp root component with streaming ([ff392d4](https://github.com/grindlemire/go-tui/commit/ff392d416d07cce98d95c41035cb883730c5015d))
* **ai-chat:** add fake provider for demo/testing ([06c006c](https://github.com/grindlemire/go-tui/commit/06c006cfd4e302ecc62aeff5783ea91c13b3d44b))
* **ai-chat:** add Header component ([038fad6](https://github.com/grindlemire/go-tui/commit/038fad6ef622b690f81101af6aa836f583675903))
* **ai-chat:** add HelpOverlay component ([8168250](https://github.com/grindlemire/go-tui/commit/8168250f057cd9b307b6ead34fe3d408ce5be2d2))
* **ai-chat:** add Message component with copy/retry actions ([4e95b3f](https://github.com/grindlemire/go-tui/commit/4e95b3f156c1e5b260cd5f1af3d32259d08a069e))
* **ai-chat:** add MessageList component with vim navigation ([e7e7950](https://github.com/grindlemire/go-tui/commit/e7e7950e1827ca81d66e92a5db2ff7d7bae18ee3))
* **ai-chat:** add provider abstraction for OpenAI/Anthropic/Ollama ([2e0cac0](https://github.com/grindlemire/go-tui/commit/2e0cac0dd7491d2cd4179ee19f24901a70fce4b9))
* **ai-chat:** add Settings screen components ([d27cf59](https://github.com/grindlemire/go-tui/commit/d27cf59fe6e96a410db9485ea41ed028e98a01f7))
* **ai-chat:** add settings screen entry point ([3e86430](https://github.com/grindlemire/go-tui/commit/3e86430ce37df748a87f9683f565ff377352917a))
* **ai-chat:** add state types and AppState ([e87e9d3](https://github.com/grindlemire/go-tui/commit/e87e9d3576597753c07f0923b06dd1632a233d54))
* **ai-chat:** wire settings screen to main app ([66fe6ad](https://github.com/grindlemire/go-tui/commit/66fe6ad8872c35dfd0cda5ec6101a5845ae6b1f7))
* **ai-chat:** wire up ChatApp with provider detection ([d587eb5](https://github.com/grindlemire/go-tui/commit/d587eb52af8326ed1c21cfbcd3d3141076267548))
* **element:** add integration tests and update dashboard example ([326e7f9](https://github.com/grindlemire/go-tui/commit/326e7f91eb72b57556cba0ec21b74fb8c9952b9d))
* **element:** add onUpdate hook for pre-render callbacks ([551fbd2](https://github.com/grindlemire/go-tui/commit/551fbd2f233cff43677a6658b4ad8ef52abeeee1))
* **element:** implement Phase 1 - Layout interface and Element core ([a3211f4](https://github.com/grindlemire/go-tui/commit/a3211f4a05c6ec048e345c3f0352939c231043be))
* **examples:** scaffold ai-chat example ([c8912ef](https://github.com/grindlemire/go-tui/commit/c8912ef4aa4562312295182208afd19597cafda6))
* integrate WatcherProvider watchers into app lifecycle ([cbc588f](https://github.com/grindlemire/go-tui/commit/cbc588f866b7432700eca499e1e5b6462647c87d))
* restore EventInspector with event tracking in interactive example ([ffc4774](https://github.com/grindlemire/go-tui/commit/ffc47741ebf6dc73f7509bc90c9891d73fce7d38))
* **tailwind:** add validation, similarity matching, and class registry (Phase 2) ([935db99](https://github.com/grindlemire/go-tui/commit/935db99d848fe89b6942a73a9afe925a68324e34))
* **tailwind:** expand class mappings with percentages, individual sides, and flex utilities (Phase 1) ([abe07ef](https://github.com/grindlemire/go-tui/commit/abe07eff7d46f6a2d05ac6b600ef97a2cd905018))
* **tailwind:** integrate class validation into analyzer and LSP diagnostics (Phase 3) ([3e1ac96](https://github.com/grindlemire/go-tui/commit/3e1ac966230947db6b01f121d397ab7c46972056))
* **tuigen:** add named element refs syntax (#Name) - Phase 1 ([91af36a](https://github.com/grindlemire/go-tui/commit/91af36ada865970907c43c44db0e9ac034fa0b8e))
* **tuigen:** add named element refs syntax (#Name) - Phase 1 ([6e7e33a](https://github.com/grindlemire/go-tui/commit/6e7e33ad27771326de5372f516738b9d92992291))
* **tuigen:** add named element refs syntax (#Name) - Phase 1 ([0e47fa4](https://github.com/grindlemire/go-tui/commit/0e47fa44551c465b740cea0f13d03fe746785d0f))
* **tuigen:** add state detection to analyzer - Phase 3 ([17fb834](https://github.com/grindlemire/go-tui/commit/17fb8349a5d8bf028fc62d885d18719a85bb021a))
* **tui:** implement App.Run(), SetRoot(), and element handlers - Phases 2 & 3 ([ba97b9a](https://github.com/grindlemire/go-tui/commit/ba97b9a7b82735e1d18070f790cd057bc834a518))
* **tui:** implement Batch() for coalescing state updates - Phase 2 ([cee1b25](https://github.com/grindlemire/go-tui/commit/cee1b25ddda813884b748a4f049e5b6da1340a41))
* **tui:** implement dirty tracking and watcher types - Phase 1 ([162a1c0](https://github.com/grindlemire/go-tui/commit/162a1c0bc289f3b9720dd5d99d457efd83ed0eee))
* **tui:** implement State[T] reactive type with bindings - Phase 1 ([99665a1](https://github.com/grindlemire/go-tui/commit/99665a143c2e8ca0dc393a29ca4eb0f8767249e3))


### Bug Fixes

* **ai-chat:** fix help text, temperature labels, and copy handler ([3f10c71](https://github.com/grindlemire/go-tui/commit/3f10c71c830c24f3ee389d11800efbe5cec4a16a))
* **ai-chat:** remove broken go:generate directive ([7fa71b2](https://github.com/grindlemire/go-tui/commit/7fa71b27ef11149945766590e05d2f5dacebd94a))
* **ai-chat:** use explicit style attrs for dynamic styling ([52ce3e3](https://github.com/grindlemire/go-tui/commit/52ce3e39d5d2ec9144174cd33146b688e1728f7f))
* align panel heights in interactive example ([1953b39](https://github.com/grindlemire/go-tui/commit/1953b39295e702cad21b1ca604cb1cb344bc02f2))
* move textElementWithOptions/skipTextChildren to generator_element.go per spec ([4f35a8f](https://github.com/grindlemire/go-tui/commit/4f35a8fa476ca173ba09550b73797ea0554afdfb))
* prevent header from shrinking with flexShrink={0} ([ab13198](https://github.com/grindlemire/go-tui/commit/ab131986faa8e2d1243238384a4ca63371de4162))
* restore previous currentApp in Run() for nested apps ([bf07b93](https://github.com/grindlemire/go-tui/commit/bf07b93575223b8d33b8dec3ee409317cc8b1713))
* settings screen as embedded component instead of separate app ([123cf98](https://github.com/grindlemire/go-tui/commit/123cf980516d45d899738aad91b6e82b95306fc9))
* use flexGrow for proper vertical distribution in interactive example ([cc10eb1](https://github.com/grindlemire/go-tui/commit/cc10eb10b1e6549aceca1f516b903abf93a4da0d))
