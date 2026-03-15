# Changelog

## [0.6.1](https://github.com/grindlemire/go-tui/compare/v0.6.0...v0.6.1) (2026-03-15)


### Bug Fixes

* **lsp:** highlight Go keywords in expression attribute values ([b147ae7](https://github.com/grindlemire/go-tui/commit/b147ae799f44c0a7743df093599b85dfb68d8614))

## [0.6.0](https://github.com/grindlemire/go-tui/compare/v0.5.0...v0.6.0) (2026-03-15)


### Features

* add KeyMatcher interface with On/OnStop/OnFocused constructors ([99c8002](https://github.com/grindlemire/go-tui/commit/99c8002d0eef138281becdb924c558a3cc7c2002))
* add Kitty keyboard negotiation to Terminal interface ([044e266](https://github.com/grindlemire/go-tui/commit/044e266af46f56345ba033a4213a8c7e287c78a5))
* add Kitty keyboard protocol escape sequences ([56e6938](https://github.com/grindlemire/go-tui/commit/56e693807502dfad2222759e22a9caeead498409))
* add Kitty keyboard protocol for Ctrl+H/I/M disambiguation ([591fbde](https://github.com/grindlemire/go-tui/commit/591fbde4cffe63a7c765434138cbc649076847cc))
* add KittyKeyboard capability and WithLegacyKeyboard option ([19cb4fc](https://github.com/grindlemire/go-tui/commit/19cb4fcc70be5d19e71f4d04bf668d3ac62ed6b5))
* implement Kitty keyboard protocol negotiation in ANSITerminal ([cac462e](https://github.com/grindlemire/go-tui/commit/cac462e2b262374c85fc298326f4d47f86e52411))
* include kitty-keyboard in capabilities string ([4a55000](https://github.com/grindlemire/go-tui/commit/4a550005463195d96528206e9900bd9860c57ae4))
* parse Kitty keyboard protocol CSI u sequences ([4f48121](https://github.com/grindlemire/go-tui/commit/4f48121211c9c5d565526e713a19c0f79ff9d742))
* wire Kitty keyboard negotiation into app lifecycle ([5d7aaaa](https://github.com/grindlemire/go-tui/commit/5d7aaaa208c37db6b273f09e9aeb8ce7525c4c3c))


### Bug Fixes

* add trailing newline to reader_types.go and document Rune(0) behavior ([fb52f4e](https://github.com/grindlemire/go-tui/commit/fb52f4e107dd92fa34f19b374129186722e3c9db))
* address code review findings for Kitty keyboard and terminal flush ([cdddab8](https://github.com/grindlemire/go-tui/commit/cdddab89007da25f2ac07e26a161144b5b83434a))
* address code review findings for Kitty keyboard implementation ([1f9def9](https://github.com/grindlemire/go-tui/commit/1f9def97f96e8c608919802d611828acb8c26907))
* handle Kitty keyboard on suspend/resume and clean up minor issues ([6f54fa4](https://github.com/grindlemire/go-tui/commit/6f54fa4c46b688c34da5a1435bb98337118506d6))
* make KeyCtrlH/I/M aliases for KeyBackspace/Tab/Enter ([4decb9a](https://github.com/grindlemire/go-tui/commit/4decb9a8d987ac9b427b434e8ce0b904cea58513))
* remove flag-2+ dead code from kittySpecialKeys and harden Kitty negotiation ([1dd9c19](https://github.com/grindlemire/go-tui/commit/1dd9c19b51443462d6d6c201caaeed077729378e))
* struct alignment via gofmt, smarter response terminator detection ([40551c0](https://github.com/grindlemire/go-tui/commit/40551c0c993de345b0a004fbca121aeb681e1b6b))

## [0.5.0](https://github.com/grindlemire/go-tui/compare/v0.4.0...v0.5.0) (2026-03-14)


### Features

* drop @ prefix from control flow keywords ([470db04](https://github.com/grindlemire/go-tui/commit/470db04495a876cf4b4f52e9527151572353545b))
* extend LetBinding AST for component call and expression RHS ([888a780](https://github.com/grindlemire/go-tui/commit/888a78008d3c5096cc0876ee79e9396ed0d4d229))
* formatter emits bare if/for/else and := instead of @-prefixed syntax ([8aefebe](https://github.com/grindlemire/go-tui/commit/8aefebe396583b48bf0b11a362ad7c6cd74487e2))
* generator handles component call and expression RHS in LetBinding ([9d68ba0](https://github.com/grindlemire/go-tui/commit/9d68ba0b348f1952be6262d5470a79ac87ab62a7))
* lexer and parser accept bare if/for/else, remove TokenAtIf/TokenAtFor/TokenAtElse ([8560391](https://github.com/grindlemire/go-tui/commit/85603914c6a8ceedee4faad7d6bc7e57edbd1bd7))
* LSP uses bare keyword syntax in completions, hover, and semantic tokens ([5cfc3de](https://github.com/grindlemire/go-tui/commit/5cfc3de4eef931db4d4e60b00be8ade71b13581d))
* parse bare if/for/else and := / var bindings with element/component RHS ([55a8047](https://github.com/grindlemire/go-tui/commit/55a8047e0b79609ccce8705345a7784013465b02))
* support multi-line component call arguments ([761cd7f](https://github.com/grindlemire/go-tui/commit/761cd7f110d7e0cb649d36710ca67718b41e457c))
* TextMate grammar highlights bare if/for/else and := bindings ([27d98e7](https://github.com/grindlemire/go-tui/commit/27d98e733a7602af65fa104221733c91b5d1d169))
* tree-sitter grammar supports bare if/for/else and := bindings ([80764b3](https://github.com/grindlemire/go-tui/commit/80764b350dfb3a6b06a7c59965781182252b9460))


### Bug Fixes

* analyzer handles component call and expression RHS in LetBinding ([3d86741](https://github.com/grindlemire/go-tui/commit/3d8674108868cf0c05173e77df1e6bdd6c88f755))
* **lsp:** drop @ prefix from offset calculations in LSP providers ([feecc14](https://github.com/grindlemire/go-tui/commit/feecc1491cd54c0ffdabadf0def6233e20cd18f5))
* **lsp:** prevent findElseKeyword false-positive on else inside strings ([394494e](https://github.com/grindlemire/go-tui/commit/394494ed848a16979423083edad963c380e71c9c))
* parseGoStatement stops at closing brace at depth 0 ([e83980f](https://github.com/grindlemire/go-tui/commit/e83980f40ac0ebd623b9ea920197168d2522ca67))
* **parser:** prevent speculative parsing from leaking errors and comments ([62caa56](https://github.com/grindlemire/go-tui/commit/62caa569d25c9f4b238358aeeb227fbd9b098a85))
* **parser:** use rune-aware iteration for UTF-8 correctness ([f542956](https://github.com/grindlemire/go-tui/commit/f54295637cabc6e2abc90cd417ff2bae77f27082))
* replace HTML-style comments with Go-style comments in docs code blocks ([1c287fe](https://github.com/grindlemire/go-tui/commit/1c287fe6c201be3295dc6baf7eee0b9d3d8f9d12))

## [0.4.0](https://github.com/grindlemire/go-tui/compare/v0.3.1...v0.4.0) (2026-03-13)


### Features

* **examples:** add directory tree data model and flatten logic ([02912e2](https://github.com/grindlemire/go-tui/commit/02912e23e3b32fd7b943a7a923272a1264ca41eb))
* **examples:** add directory tree example ([186b2d0](https://github.com/grindlemire/go-tui/commit/186b2d0d934b54bb20fda8ef8d9d3904174cfb8f))
* **examples:** add directory tree example skeleton ([639da53](https://github.com/grindlemire/go-tui/commit/639da53217172ebcfa9f25ac72010dc7b2e818be))
* **examples:** add directory tree keyboard navigation ([1bdae4e](https://github.com/grindlemire/go-tui/commit/1bdae4ee1adb027508d790f6ebde905e6bb17b8f))
* **examples:** add directory tree render and complete example ([480566e](https://github.com/grindlemire/go-tui/commit/480566e5fb2c281e6912fe7c853af76338a0103a))
* **examples:** add scrolling, overflow clipping, and ancestor path highlighting ([ee46208](https://github.com/grindlemire/go-tui/commit/ee46208ea46bf57d54ce444073a3286a4bbe7e11))
* **examples:** replace hardcoded tree with random generator ([b5aaa4b](https://github.com/grindlemire/go-tui/commit/b5aaa4b88f982231677779b012777c78c7d3b421))
* **examples:** show selected node path at top of directory tree ([4d7f0ee](https://github.com/grindlemire/go-tui/commit/4d7f0ee4f30d6ec0982cd2771ca2d5ed177a836c))


### Bug Fixes

* **docs:** move Google Fonts from CSS [@import](https://github.com/import) to HTML link tag ([e0ee0d0](https://github.com/grindlemire/go-tui/commit/e0ee0d0e8f9d7988875da85df4ae6e1d5d6a8670))
* **docs:** move playwright to devDependencies and fix tview widget count ([269e858](https://github.com/grindlemire/go-tui/commit/269e8588fffafff8b3906af4fdf5fdfdef07eab0))
* **docs:** remove unused useRef import ([7847f15](https://github.com/grindlemire/go-tui/commit/7847f157275d34c6827937d73e67bbcf02274bcb))
* **examples:** remove overflow-hidden that bypassed scrollable rendering ([39ea661](https://github.com/grindlemire/go-tui/commit/39ea6617256cfe8c32beee508cbd97f1af9f1fe2))
* **examples:** use state-driven scrollOffset for directory tree scrolling ([a272a91](https://github.com/grindlemire/go-tui/commit/a272a912b0e115f721a6684535ecc5eb782623d1))
* **lsp:** use character offsets for semantic tokens ([b7bcf07](https://github.com/grindlemire/go-tui/commit/b7bcf07653120b43d3b781bee7a3f2e4bdf4bc9a))
* **lsp:** use character offsets instead of byte offsets for semantic tokens ([1791cf6](https://github.com/grindlemire/go-tui/commit/1791cf6d757e891dce01719ef4ef75c0c3874e73))

## [0.3.1](https://github.com/grindlemire/go-tui/compare/v0.3.0...v0.3.1) (2026-03-09)


### Bug Fixes

* move SIGWINCH handling from reader to App and default to blocking input ([3b2879a](https://github.com/grindlemire/go-tui/commit/3b2879ab99fbbfae8fda2360df1041bef2db6035))

## [0.3.0](https://github.com/grindlemire/go-tui/compare/v0.2.0...v0.3.0) (2026-03-09)


### Features

* add autoFocus attribute and make focusColor optional for Input and TextArea ([184af94](https://github.com/grindlemire/go-tui/commit/184af94620528453fd9615d39959951d0d7fdd9a))
* add borderGradient, focusGradient, and backspace scroll fix for Input and TextArea ([d1437fd](https://github.com/grindlemire/go-tui/commit/d1437fd0a85cece58ad68bc7b7071c64861255ca))
* add FocusRequired flag to KeyPattern and focus-gated binding helpers ([5d97ee1](https://github.com/grindlemire/go-tui/commit/5d97ee153fabb842690fcb72c67e2c6fd768220a))
* add IsFocused query method to focusManager ([681ffb3](https://github.com/grindlemire/go-tui/commit/681ffb3e3fd450a1b56be15b8c38e0d440e22682))
* add reactive value binding and focusColor for Input and TextArea ([35d1679](https://github.com/grindlemire/go-tui/commit/35d167916c9058f6f89d822d3b1d1cfc127a32f7))
* implement focus-gated dispatch in dispatch table ([0ea047a](https://github.com/grindlemire/go-tui/commit/0ea047ad18050e78f966a00c955134e9875875b2))
* separate tabStop from focusable to fix Tab navigation ([4d6189d](https://github.com/grindlemire/go-tui/commit/4d6189d96d83ed1cc19c4cb84668c0283a75bf00))
* wire Input component into focus system with focus-gated key bindings ([d2689fa](https://github.com/grindlemire/go-tui/commit/d2689fa29436bee0fb06aecc871e0f0d7dc2502c))
* wire TextArea component into focus system with focus-gated key bindings ([af57b8f](https://github.com/grindlemire/go-tui/commit/af57b8fdc92008c6eb19a780567d7739cd3a253c))


### Bug Fixes

* add Tab/Shift+Tab focus navigation to elements example ([7f77e38](https://github.com/grindlemire/go-tui/commit/7f77e38f410b34e36671a2345ab932e41ac3bdc1))
* exclude focus-gated entries from dispatch table conflict validation ([1545aa6](https://github.com/grindlemire/go-tui/commit/1545aa6c51ecf40f7c6ff69fa35a1d80656f5901))
* make ContainsPoint account for scroll offset in scrollable containers ([23be70e](https://github.com/grindlemire/go-tui/commit/23be70eebcd4ba8a04961fb361f6c908ec540717))

## [0.2.0](https://github.com/grindlemire/go-tui/compare/v0.1.2...v0.2.0) (2026-03-07)


### Features

* add default Ctrl+Z suspend fallback in key dispatch ([8f61d4e](https://github.com/grindlemire/go-tui/commit/8f61d4e161a5815f3ce7d267d20fb6be256eea65))
* add FlexWrap and AlignContent types with public API and Tailwind classes ([28eb41f](https://github.com/grindlemire/go-tui/commit/28eb41fedb83b8ac7f53a72399b03fa31524a11f))
* add no-op suspend stubs for Windows ([0484abf](https://github.com/grindlemire/go-tui/commit/0484abff025240f9190e50912fdc2912f07d2dcb))
* add OnChange watcher for reactive state effects ([623b93f](https://github.com/grindlemire/go-tui/commit/623b93fbb7a0e21db4bbdfa8d5667dd8cdd5c652))
* add OnKeyMod helper, fix flex-wrap align-content, update docs ([dbcce85](https://github.com/grindlemire/go-tui/commit/dbcce85edfed32c2780be63fd4de5d9a16f74c8b))
* add onSuspend/onResume fields to App struct ([f8be019](https://github.com/grindlemire/go-tui/commit/f8be019ef4ece3f4f4cbebfda5ef30cd0cb765b8))
* add WithOnSuspend and WithOnResume app options ([610b3ad](https://github.com/grindlemire/go-tui/commit/610b3ad76d69c927834eea455d42e4d541013ce1))
* implement suspend/resume terminal state management ([699e343](https://github.com/grindlemire/go-tui/commit/699e3435a3880e77cc2e2e59ada4ca4d1fab415c))
* **layout:** implement flex-wrap line breaking, per-line layout, and auto cross-axis sizing ([6932cd5](https://github.com/grindlemire/go-tui/commit/6932cd57dfd31eafa8b420655d05c5add0903ed3))
* register SIGTSTP signal handler in app event loop ([8e7163e](https://github.com/grindlemire/go-tui/commit/8e7163ee52d6321c6a40cbf70244ead99e007d12))


### Bug Fixes

* bake inline widget to scrollback before suspend ([f076cb3](https://github.com/grindlemire/go-tui/commit/f076cb369e41b3a97cb06af234feb393951902dc))
* clear widget area on inline suspend instead of baking duplicate ([a1ee2c2](https://github.com/grindlemire/go-tui/commit/a1ee2c247e50f6745f719aa72781195a2ec4df48))
* handle inline mode suspend/resume without corrupting scrollback ([03c1aae](https://github.com/grindlemire/go-tui/commit/03c1aae3574974b5b3851c9a02f0a64705bea6a4))
* prevent stack overflow from circular state dependencies ([f1c9ba5](https://github.com/grindlemire/go-tui/commit/f1c9ba58ec1e81b851f307319f6af8b83cbdc0b4))
* re-register SIGTSTP signal handler after resume ([554fded](https://github.com/grindlemire/go-tui/commit/554fdedbad37f379f92da0547bb454d94c79ec6b))
* replace sleep-based test sync with done channel, document FlexGrow heuristic ([dcc95f0](https://github.com/grindlemire/go-tui/commit/dcc95f05ad8d0aef43f5eac90b4b53af8debc1df))
* resolve three suspend/resume issues ([7e5b061](https://github.com/grindlemire/go-tui/commit/7e5b0614c918c873187891d0433ddafe17800185))

## [0.1.2](https://github.com/grindlemire/go-tui/compare/v0.1.1...v0.1.2) (2026-03-04)


### Bug Fixes

* use runtime/debug to report version from go install ([bd586f8](https://github.com/grindlemire/go-tui/commit/bd586f84e8c465e20f4d36b980721d077c27ec4b))

## [0.1.1](https://github.com/grindlemire/go-tui/compare/v0.1.0...v0.1.1) (2026-03-04)


### Bug Fixes

* update README badges, examples, and VS Code extension docs ([8c9fdf4](https://github.com/grindlemire/go-tui/commit/8c9fdf444d7bdbcd29633b7b7c1aedb249cd62d5))

## 0.1.0 (2026-03-03)


### Features

* add 1-character minimum spacing between table columns ([0036fc0](https://github.com/grindlemire/go-tui/commit/0036fc01828fb4983b96ce554766b8d488e2237e))
* add ANSI-aware byte scanner for styled streaming ([37c51d3](https://github.com/grindlemire/go-tui/commit/37c51d36a06b35717167baab960bcecc10f4e78f))
* add App.StreamAbove() with PrintAbove coordination ([09f6eac](https://github.com/grindlemire/go-tui/commit/09f6eacc14675fcef1d804c7dca024c8aae68e75))
* add bufferRowToANSI for rendering buffer rows to ANSI strings ([a2c1db9](https://github.com/grindlemire/go-tui/commit/a2c1db96dbb11b34ae820c89f376b14276f335cb))
* add ClickBinding type for ref-based mouse handling ([a18bf5b](https://github.com/grindlemire/go-tui/commit/a18bf5b0b90f47e4e807845db9e9c5d41031fb44))
* add Element.Component() public getter ([2c489a5](https://github.com/grindlemire/go-tui/commit/2c489a501e4d76439a8691221df91137c0a9b5d9))
* add ElementTag field for table layout dispatch ([51daa21](https://github.com/grindlemire/go-tui/commit/51daa21251c0cb845300f8b06cce6fbadc574176))
* add generic NewChannelWatcher helper ([52f3f86](https://github.com/grindlemire/go-tui/commit/52f3f86082ef2285a04f5b684ada9ceabf92edc7))
* add HandleClicks helper for automatic ref hit testing ([ba91e9d](https://github.com/grindlemire/go-tui/commit/ba91e9d9d6b79c06cb16eb6198ad635450bfa5fa))
* add HeightForWidth to Layoutable for text wrap height calculation ([5edcd48](https://github.com/grindlemire/go-tui/commit/5edcd484f554040cf5cd56feefef6a11adb3a752))
* add inlineStreamWriter and nopStreamWriter types ([e180a59](https://github.com/grindlemire/go-tui/commit/e180a594f8982cdaf7fad109ca760be53e6784b6))
* add MountPersistent to prevent component sweep when hidden ([0b8ce63](https://github.com/grindlemire/go-tui/commit/0b8ce6345c1edaa5a97fc1fa7b75e074bc2f7c38))
* add nowrap and wrap tailwind classes ([43e4b5a](https://github.com/grindlemire/go-tui/commit/43e4b5ae72b30bdf724627b33692a6c798a09939))
* add partial line tracking and appendBytes to inlineSession ([cb55fd3](https://github.com/grindlemire/go-tui/commit/cb55fd35a4a2c7ab937c75e782541561e9540bf9))
* add Print, Sprint, and Fprint for single-frame rendering ([c7b4660](https://github.com/grindlemire/go-tui/commit/c7b46603647f8147cf4f4d02e30807eb33185a2e))
* add Print, Sprint, and Fprint for single-frame rendering ([6c580b5](https://github.com/grindlemire/go-tui/commit/6c580b58ad1fb8df01cf8cb0cf3f428dbef78ae9))
* add PrintAboveElement and QueuePrintAboveElement for inline mode ([60e23c5](https://github.com/grindlemire/go-tui/commit/60e23c520dfae9131bb9ba63f78eccf06b537480))
* add renderElementToBuffer for standalone element rendering ([2899fc6](https://github.com/grindlemire/go-tui/commit/2899fc6bdcf838d51f2515ca3cbdfb00ca3dfd4d))
* add single-frame print example (examples/19-print) ([8dd1aef](https://github.com/grindlemire/go-tui/commit/8dd1aeffed882f9762b4ce3449190e46776de0f0))
* add StreamWriter.WriteElement for mid-stream element insertion ([fc7e4ae](https://github.com/grindlemire/go-tui/commit/fc7e4ae92c38c50d111d693297daf38504a769f5))
* add tree walk to collect watchers from WatcherProvider components ([2abe659](https://github.com/grindlemire/go-tui/commit/2abe659c861332e3f8cc3772c2f1e4b3d88244f6))
* add WatcherProvider interface for component-level watchers ([4ceac54](https://github.com/grindlemire/go-tui/commit/4ceac54f40f893b1531a4f83adc69a69a409974f))
* add WithWrap option for text wrapping (default enabled) ([8efdf97](https://github.com/grindlemire/go-tui/commit/8efdf97a0d5502787682847140d3120846160d9a))
* add word-wrap text function with mid-word fallback ([2632126](https://github.com/grindlemire/go-tui/commit/263212657d99ff948601d0d81ef6ddca184f707d))
* add Wrap() public getter to Element ([448c1cd](https://github.com/grindlemire/go-tui/commit/448c1cde54dfdd054ae8a00d2e4efedefa2c4809))
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
* **ai-chat:** migrate textarea to GSX element with persistent mount ([f83b016](https://github.com/grindlemire/go-tui/commit/f83b0169e1afe734391909703df911c2144f305e))
* **ai-chat:** wire settings screen to main app ([66fe6ad](https://github.com/grindlemire/go-tui/commit/66fe6ad8872c35dfd0cda5ec6101a5845ae6b1f7))
* **ai-chat:** wire up ChatApp with provider detection ([d587eb5](https://github.com/grindlemire/go-tui/commit/d587eb52af8326ed1c21cfbcd3d3141076267548))
* **analyzer:** accept textarea as valid element tag ([0fdb81a](https://github.com/grindlemire/go-tui/commit/0fdb81abf08a2b313db6c5970a8ed7b7a75396f4))
* auto-scroll wrapped text with hidden scrollbar on overflow ([b3fa83c](https://github.com/grindlemire/go-tui/commit/b3fa83c0db75d03cbdbc7e2a63f6ce5c0d7af983))
* compute table intrinsic size from column widths and row heights ([b5b9634](https://github.com/grindlemire/go-tui/commit/b5b963464633edb69ed6dc1db5cb56a60b767371))
* **element:** add integration tests and update dashboard example ([326e7f9](https://github.com/grindlemire/go-tui/commit/326e7f91eb72b57556cba0ec21b74fb8c9952b9d))
* **element:** add onUpdate hook for pre-render callbacks ([551fbd2](https://github.com/grindlemire/go-tui/commit/551fbd2f233cff43677a6658b4ad8ef52abeeee1))
* **element:** implement Phase 1 - Layout interface and Element core ([a3211f4](https://github.com/grindlemire/go-tui/commit/a3211f4a05c6ec048e345c3f0352939c231043be))
* **examples:** scaffold ai-chat example ([c8912ef](https://github.com/grindlemire/go-tui/commit/c8912ef4aa4562312295182208afd19597cafda6))
* generate WithTag for table elements in codegen ([a854717](https://github.com/grindlemire/go-tui/commit/a854717a45054d331cddc8d94f74341ef676b435))
* **generator:** emit app.Mount + NewTextArea for textarea elements ([31c29d7](https://github.com/grindlemire/go-tui/commit/31c29d7031f9760c011170b460fd54be1b7ebba3))
* **generator:** emit MountPersistent for component elements ([2635176](https://github.com/grindlemire/go-tui/commit/26351768a47f326ecb8f9fa7b71516323b577218))
* implement table layout algorithm with auto column sizing ([5d1d76a](https://github.com/grindlemire/go-tui/commit/5d1d76af49d1cac71dfa016221223d33d970b8e9))
* integrate WatcherProvider watchers into app lifecycle ([cbc588f](https://github.com/grindlemire/go-tui/commit/cbc588f866b7432700eca499e1e5b6462647c87d))
* register tr, td, th elements in schema and analyzer ([20a41d9](https://github.com/grindlemire/go-tui/commit/20a41d98b9dc1970b1900f45294e8b286bdccbb9))
* render th elements with bold text by default ([c7a0849](https://github.com/grindlemire/go-tui/commit/c7a084945f86e8503370bb893ea496383c6fbec8))
* render wrapped text across multiple lines with per-line alignment ([12df559](https://github.com/grindlemire/go-tui/commit/12df5598b14c05baf254fd2b735b69e3891c9029))
* restore EventInspector with event tracking in interactive example ([ffc4774](https://github.com/grindlemire/go-tui/commit/ffc47741ebf6dc73f7509bc90c9891d73fce7d38))
* **schema:** add textarea element definition to LSP schema ([d141b42](https://github.com/grindlemire/go-tui/commit/d141b4277adccd5e158ae13e45089fe6c0641bdb))
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
* two-pass layout for text wrapping height calculation ([afb032f](https://github.com/grindlemire/go-tui/commit/afb032f3e5435c331a7f9c015e563e7cd51118f9))


### Bug Fixes

* add explicit row height override for table tr elements ([0f50303](https://github.com/grindlemire/go-tui/commit/0f50303b24e5a7c047bc3aad32aae59277c8e3eb))
* **ai-chat:** fix help text, temperature labels, and copy handler ([3f10c71](https://github.com/grindlemire/go-tui/commit/3f10c71c830c24f3ee389d11800efbe5cec4a16a))
* **ai-chat:** remove broken go:generate directive ([7fa71b2](https://github.com/grindlemire/go-tui/commit/7fa71b27ef11149945766590e05d2f5dacebd94a))
* **ai-chat:** use explicit style attrs for dynamic styling ([52ce3e3](https://github.com/grindlemire/go-tui/commit/52ce3e39d5d2ec9144174cd33146b688e1728f7f))
* align panel heights in interactive example ([1953b39](https://github.com/grindlemire/go-tui/commit/1953b39295e702cad21b1ca604cb1cb344bc02f2))
* flaky tests, nil-ref panic, and fill test coverage gaps ([9c3491c](https://github.com/grindlemire/go-tui/commit/9c3491c8bbc8aa29b2bfe847385047a5737c28f8))
* guard against nil ref in HandleClicks ([2380913](https://github.com/grindlemire/go-tui/commit/238091376131c172900a0fac1bdfb31b0fd34609))
* move textElementWithOptions/skipTextChildren to generator_element.go per spec ([4f35a8f](https://github.com/grindlemire/go-tui/commit/4f35a8fa476ca173ba09550b73797ea0554afdfb))
* prevent header from shrinking with flexShrink={0} ([ab13198](https://github.com/grindlemire/go-tui/commit/ab131986faa8e2d1243238384a4ca63371de4162))
* recursive HeightForWidth for containers to propagate wrapped text height ([28c3486](https://github.com/grindlemire/go-tui/commit/28c34865922157777b8d6225d4952069c446761f))
* replace time.Sleep with channel draining in watcher tests ([994d0cb](https://github.com/grindlemire/go-tui/commit/994d0cb96bfbda8da942695ba5c062abfe1d46b5))
* restore previous currentApp in Run() for nested apps ([bf07b93](https://github.com/grindlemire/go-tui/commit/bf07b93575223b8d33b8dec3ee409317cc8b1713))
* settings screen as embedded component instead of separate app ([123cf98](https://github.com/grindlemire/go-tui/commit/123cf980516d45d899738aad91b6e82b95306fc9))
* use flexGrow for proper vertical distribution in interactive example ([cc10eb1](https://github.com/grindlemire/go-tui/commit/cc10eb10b1e6549aceca1f516b903abf93a4da0d))
