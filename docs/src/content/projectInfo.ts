export const projectInfo = {
  name: "go-tui",
  tagline: "Reactive Terminal UIs in Go",
  description:
    "A Go framework for terminal UIs. Define layout in .gsx templates with HTML-like syntax, compile to type-safe Go. Flexbox positioning, reactive state, no external dependencies.",
  installCmd: "go get github.com/grindlemire/go-tui",
  features: [
    {
      title: "Declarative Syntax",
      description:
        ".gsx templates use HTML-like elements and Tailwind-style classes. The compiler outputs type-safe Go. No runtime reflection, no magic strings.",
      icon: "code",
    },
    {
      title: "Pure Go Flexbox",
      description:
        "Row, column, justify, align, gap, padding, margin \u2014 real flexbox layout without manual coordinate math. No external dependencies.",
      icon: "layout",
    },
    {
      title: "Reactive State",
      description:
        "Generic State[T] re-renders automatically on change. Bind callbacks and batch updates.",
      icon: "zap",
    },
    {
      title: "Component System",
      description:
        "Components accept parameters, refs, and watchers. Keyboard and mouse events are handled per-component.",
      icon: "box",
    },
    {
      title: "Editor Support",
      description:
        "Language server, formatter, and tree-sitter grammar for syntax highlighting, completions, go-to-definition, and more.",
      icon: "edit",
    },
    {
      title: "Minimal Dependencies",
      description:
        "Only golang.org/x standard libraries. No external dependencies.",
      icon: "package",
    },
  ],
};

export const tailwindClasses = [
  { class: "flex", description: "Display flex (row direction)" },
  { class: "flex-col", description: "Display flex (column direction)" },
  { class: "grow", description: "Flex grow to fill space" },
  { class: "shrink-0", description: "Prevent flex shrink" },
  { class: "gap-N", description: "Gap of N characters between children" },
  { class: "p-N", description: "Padding of N on all sides" },
  { class: "px-N", description: "Horizontal padding of N" },
  { class: "py-N", description: "Vertical padding of N" },
  { class: "m-N", description: "Margin of N on all sides" },
  { class: "border-single", description: "Single line border \u250C\u2500\u2510\u2502\u2514\u2500\u2518" },
  { class: "border-double", description: "Double line border \u2554\u2550\u2557\u2551\u2558\u2550\u255D" },
  { class: "border-rounded", description: "Rounded border \u256D\u2500\u256E\u2502\u2570\u2500\u256F" },
  { class: "border-thick", description: "Thick border \u250F\u2501\u2513\u2503\u2517\u2501\u251B" },
  { class: "font-bold", description: "Bold text" },
  { class: "font-dim", description: "Dim/faint text" },
  { class: "font-italic", description: "Italic text" },
  { class: "font-underline", description: "Underlined text" },
  { class: "text-COLOR", description: "Text color (red, green, blue, cyan, magenta, yellow, white)" },
  { class: "bg-COLOR", description: "Background color" },
  { class: "items-center", description: "Align items center (cross axis)" },
  { class: "items-start", description: "Align items start" },
  { class: "items-end", description: "Align items end" },
  { class: "items-stretch", description: "Stretch items to fill" },
  { class: "justify-center", description: "Justify content center (main axis)" },
  { class: "justify-between", description: "Space between items" },
  { class: "justify-around", description: "Space around items" },
  { class: "justify-evenly", description: "Space evenly between items" },
  { class: "h-full", description: "Height 100%" },
  { class: "w-full", description: "Width 100%" },
  { class: "text-center", description: "Center text alignment" },
  { class: "text-right", description: "Right text alignment" },
  { class: "truncate", description: "Truncate overflowing text with ellipsis" },
];
