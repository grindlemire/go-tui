// Package schema provides centralized language knowledge for the GSX LSP.
// All element definitions, attribute types, and event handlers are defined here
// as a single source of truth used by all LSP providers.
package schema

// ElementDef describes a built-in GSX element.
type ElementDef struct {
	Tag         string
	Description string
	Attributes  []AttributeDef
	SelfClosing bool
	Category    string // "container", "text", "input", "display"
}

// AttributeDef describes an element attribute.
type AttributeDef struct {
	Name        string
	Type        string // "string", "int", "float", "bool", "expression", "direction", "justify", "align", "border", "color", "style", "func"
	Description string
	Category    string // "layout", "visual", "event", "ref", "spacing", "flex", "text", "scroll", "generic"
}

// EventHandlerDef describes an event handler attribute.
type EventHandlerDef struct {
	Name        string
	Description string
	Signature   string // Expected handler signature
}

// Elements maps tag names to their definitions.
var Elements = map[string]*ElementDef{
	"div": {
		Tag:         "div",
		Description: "A block container with flexbox layout. The primary building block for layouts.",
		Attributes:  containerAttrs(),
		Category:    "container",
	},
	"span": {
		Tag:         "span",
		Description: "An inline text container for styling text content.",
		Attributes:  textElementAttrs(),
		Category:    "text",
	},
	"p": {
		Tag:         "p",
		Description: "A paragraph element for text blocks.",
		Attributes:  textElementAttrs(),
		Category:    "text",
	},
	"ul": {
		Tag:         "ul",
		Description: "An unordered list container. Use with `<li>` children.",
		Attributes:  containerAttrs(),
		Category:    "container",
	},
	"li": {
		Tag:         "li",
		Description: "A list item. Should be a child of `<ul>`.",
		Attributes:  containerAttrs(),
		Category:    "container",
	},
	"button": {
		Tag:         "button",
		Description: "A clickable button element that can receive focus and handle events.",
		Attributes:  buttonAttrs(),
		Category:    "input",
	},
	"input": {
		Tag:         "input",
		Description: "A text input field for user input.",
		Attributes:  inputAttrs(),
		SelfClosing: true,
		Category:    "input",
	},
	"table": {
		Tag:         "table",
		Description: "A table container for tabular data.",
		Attributes:  containerAttrs(),
		Category:    "display",
	},
	"progress": {
		Tag:         "progress",
		Description: "A progress bar element showing completion status.",
		Attributes:  progressAttrs(),
		SelfClosing: true,
		Category:    "display",
	},
	"hr": {
		Tag:         "hr",
		Description: "A horizontal dividing line.",
		SelfClosing: true,
		Category:    "display",
		Attributes: []AttributeDef{
			{Name: "id", Type: "string", Description: "Unique identifier for the element", Category: "generic"},
			{Name: "class", Type: "string", Description: "Tailwind-style CSS classes", Category: "generic"},
		},
	},
	"br": {
		Tag:         "br",
		Description: "An empty line break.",
		SelfClosing: true,
		Category:    "display",
		Attributes: []AttributeDef{
			{Name: "id", Type: "string", Description: "Unique identifier for the element", Category: "generic"},
			{Name: "class", Type: "string", Description: "Tailwind-style CSS classes", Category: "generic"},
		},
	},
}

// EventHandlers maps event attribute names to their definitions.
var EventHandlers = map[string]*EventHandlerDef{
	"onFocus": {
		Name:        "onFocus",
		Description: "Called when the element gains focus.",
		Signature:   "func()",
	},
	"onBlur": {
		Name:        "onBlur",
		Description: "Called when the element loses focus.",
		Signature:   "func()",
	},
	"onChannel": {
		Name:        "onChannel",
		Description: "Called when a message is received on the channel watcher.",
		Signature:   "func()",
	},
	"onTimer": {
		Name:        "onTimer",
		Description: "Called on each timer tick.",
		Signature:   "func()",
	},
}

// VoidElements returns the set of elements that cannot have children.
// This is derived from Elements[tag].SelfClosing to maintain a single source of truth.
func VoidElements() map[string]bool {
	void := make(map[string]bool)
	for tag, elem := range Elements {
		if elem.SelfClosing {
			void[tag] = true
		}
	}
	return void
}

// IsVoidElement returns true if the tag is a void element (cannot have children).
func IsVoidElement(tag string) bool {
	elem := Elements[tag]
	return elem != nil && elem.SelfClosing
}

// GetElement returns the definition for a tag, or nil if unknown.
func GetElement(tag string) *ElementDef {
	return Elements[tag]
}

// GetAttribute returns the attribute definition for a given tag and attribute name,
// or nil if not found.
func GetAttribute(tag, attr string) *AttributeDef {
	elem := Elements[tag]
	if elem == nil {
		return nil
	}
	for i := range elem.Attributes {
		if elem.Attributes[i].Name == attr {
			return &elem.Attributes[i]
		}
	}
	return nil
}

// GetEventHandler returns the event handler definition, or nil if unknown.
func GetEventHandler(name string) *EventHandlerDef {
	return EventHandlers[name]
}

// IsElementTag returns true if the tag is a known built-in element.
func IsElementTag(tag string) bool {
	_, ok := Elements[tag]
	return ok
}

// IsEventHandler returns true if the attribute name is an event handler.
func IsEventHandler(name string) bool {
	_, ok := EventHandlers[name]
	return ok
}

// AllElementTags returns all known element tag names.
func AllElementTags() []string {
	tags := make([]string, 0, len(Elements))
	for tag := range Elements {
		tags = append(tags, tag)
	}
	return tags
}

// AllEventHandlerNames returns all known event handler attribute names.
func AllEventHandlerNames() []string {
	names := make([]string, 0, len(EventHandlers))
	for name := range EventHandlers {
		names = append(names, name)
	}
	return names
}

// --- Attribute sets ---

func genericAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "id", Type: "string", Description: "Unique identifier for the element", Category: "generic"},
		{Name: "class", Type: "string", Description: "Tailwind-style CSS classes", Category: "generic"},
		{Name: "disabled", Type: "bool", Description: "Whether the element is disabled", Category: "generic"},
		{Name: "deps", Type: "expression", Description: "Explicit state dependencies for reactive bindings", Category: "generic"},
		{Name: "ref", Type: "expression", Description: "Bind this element to a ref variable (tui.NewRef/NewRefList/NewRefMap)", Category: "ref"},
	}
}

func layoutAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "width", Type: "int", Description: "Fixed width in characters", Category: "layout"},
		{Name: "widthPercent", Type: "int", Description: "Width as percentage of parent", Category: "layout"},
		{Name: "height", Type: "int", Description: "Fixed height in rows", Category: "layout"},
		{Name: "heightPercent", Type: "int", Description: "Height as percentage of parent", Category: "layout"},
		{Name: "minWidth", Type: "int", Description: "Minimum width", Category: "layout"},
		{Name: "minHeight", Type: "int", Description: "Minimum height", Category: "layout"},
		{Name: "maxWidth", Type: "int", Description: "Maximum width", Category: "layout"},
		{Name: "maxHeight", Type: "int", Description: "Maximum height", Category: "layout"},
	}
}

func flexAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "direction", Type: "direction", Description: "Flex direction (row, column)", Category: "flex"},
		{Name: "justify", Type: "justify", Description: "Justify content (start, center, end, between, around)", Category: "flex"},
		{Name: "align", Type: "align", Description: "Align items (start, center, end, stretch)", Category: "flex"},
		{Name: "gap", Type: "int", Description: "Gap between children", Category: "flex"},
		{Name: "flexGrow", Type: "float", Description: "Flex grow factor", Category: "flex"},
		{Name: "flexShrink", Type: "float", Description: "Flex shrink factor", Category: "flex"},
		{Name: "alignSelf", Type: "align", Description: "Override parent's align for this item", Category: "flex"},
	}
}

func spacingAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "padding", Type: "int", Description: "Padding on all sides", Category: "spacing"},
		{Name: "margin", Type: "int", Description: "Margin on all sides", Category: "spacing"},
	}
}

func visualAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "border", Type: "border", Description: "Border style (none, single, double, rounded, thick)", Category: "visual"},
		{Name: "borderStyle", Type: "string", Description: "Border style name", Category: "visual"},
		{Name: "background", Type: "color", Description: "Background color", Category: "visual"},
	}
}

func textAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "text", Type: "string", Description: "Text content", Category: "text"},
		{Name: "textStyle", Type: "style", Description: "Text styling", Category: "text"},
		{Name: "textAlign", Type: "string", Description: "Text alignment (left, center, right)", Category: "text"},
	}
}

func eventAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "onFocus", Type: "func", Description: "Focus gained handler", Category: "event"},
		{Name: "onBlur", Type: "func", Description: "Focus lost handler", Category: "event"},
		{Name: "focusable", Type: "bool", Description: "Whether the element can receive focus", Category: "event"},
	}
}

func watcherAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "onChannel", Type: "expression", Description: "Channel watcher for async updates", Category: "event"},
		{Name: "onTimer", Type: "expression", Description: "Timer watcher for periodic updates", Category: "event"},
	}
}

func scrollAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "scrollable", Type: "bool", Description: "Enable scrolling for overflow content", Category: "scroll"},
		{Name: "scrollbarStyle", Type: "style", Description: "Style for the scrollbar track", Category: "scroll"},
		{Name: "scrollbarThumbStyle", Type: "style", Description: "Style for the scrollbar thumb", Category: "scroll"},
	}
}

// containerAttrs returns all attributes for container elements (div, ul, li, table).
func containerAttrs() []AttributeDef {
	var attrs []AttributeDef
	attrs = append(attrs, genericAttrs()...)
	attrs = append(attrs, layoutAttrs()...)
	attrs = append(attrs, flexAttrs()...)
	attrs = append(attrs, spacingAttrs()...)
	attrs = append(attrs, visualAttrs()...)
	attrs = append(attrs, eventAttrs()...)
	attrs = append(attrs, watcherAttrs()...)
	attrs = append(attrs, scrollAttrs()...)
	return attrs
}

// textElementAttrs returns attributes for text elements (span, p).
func textElementAttrs() []AttributeDef {
	var attrs []AttributeDef
	attrs = append(attrs, genericAttrs()...)
	attrs = append(attrs, textAttrs()...)
	attrs = append(attrs, layoutAttrs()...)
	attrs = append(attrs, spacingAttrs()...)
	attrs = append(attrs, visualAttrs()...)
	attrs = append(attrs, eventAttrs()...)
	return attrs
}

// buttonAttrs returns attributes for button elements.
func buttonAttrs() []AttributeDef {
	var attrs []AttributeDef
	attrs = append(attrs, genericAttrs()...)
	attrs = append(attrs, textAttrs()...)
	attrs = append(attrs, layoutAttrs()...)
	attrs = append(attrs, spacingAttrs()...)
	attrs = append(attrs, visualAttrs()...)
	attrs = append(attrs, eventAttrs()...)
	return attrs
}

// inputAttrs returns attributes for input elements.
func inputAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "id", Type: "string", Description: "Unique identifier for the element", Category: "generic"},
		{Name: "class", Type: "string", Description: "Tailwind-style CSS classes", Category: "generic"},
		{Name: "value", Type: "string", Description: "Current input value", Category: "generic"},
		{Name: "placeholder", Type: "string", Description: "Placeholder text when empty", Category: "generic"},
		{Name: "width", Type: "int", Description: "Input width in characters", Category: "layout"},
		{Name: "disabled", Type: "bool", Description: "Whether input is disabled", Category: "generic"},
		{Name: "onFocus", Type: "func", Description: "Focus gained handler", Category: "event"},
		{Name: "onBlur", Type: "func", Description: "Focus lost handler", Category: "event"},
	}
}

// progressAttrs returns attributes for progress elements.
func progressAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "id", Type: "string", Description: "Unique identifier for the element", Category: "generic"},
		{Name: "class", Type: "string", Description: "Tailwind-style CSS classes", Category: "generic"},
		{Name: "value", Type: "int", Description: "Current progress value (0 to max)", Category: "generic"},
		{Name: "max", Type: "int", Description: "Maximum progress value", Category: "generic"},
		{Name: "width", Type: "int", Description: "Progress bar width in characters", Category: "layout"},
	}
}
