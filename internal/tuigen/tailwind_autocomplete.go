package tuigen

// TailwindClassInfo contains metadata about a class for autocomplete
type TailwindClassInfo struct {
	Name        string
	Category    string // "layout", "spacing", "typography", "visual", "flex"
	Description string
	Example     string
}

// AllTailwindClasses returns all known classes for autocomplete
func AllTailwindClasses() []TailwindClassInfo {
	var classes []TailwindClassInfo

	// Layout classes
	layoutClasses := []TailwindClassInfo{
		{Name: "flex", Category: "layout", Description: "Display flex row", Example: `<div class="flex">`},
		{Name: "flex-row", Category: "layout", Description: "Display flex row", Example: `<div class="flex-row">`},
		{Name: "flex-col", Category: "layout", Description: "Display flex column", Example: `<div class="flex-col">`},
	}
	classes = append(classes, layoutClasses...)

	// Flex utilities
	flexClasses := []TailwindClassInfo{
		{Name: "flex-grow", Category: "flex", Description: "Allow element to grow", Example: `<div class="flex-grow">`},
		{Name: "flex-shrink", Category: "flex", Description: "Allow element to shrink", Example: `<div class="flex-shrink">`},
		{Name: "flex-grow-0", Category: "flex", Description: "Prevent element from growing", Example: `<div class="flex-grow-0">`},
		{Name: "flex-shrink-0", Category: "flex", Description: "Prevent element from shrinking", Example: `<div class="flex-shrink-0">`},
	}
	classes = append(classes, flexClasses...)

	// Justify content
	justifyClasses := []TailwindClassInfo{
		{Name: "justify-start", Category: "flex", Description: "Justify content to start", Example: `<div class="flex justify-start">`},
		{Name: "justify-center", Category: "flex", Description: "Justify content to center", Example: `<div class="flex justify-center">`},
		{Name: "justify-end", Category: "flex", Description: "Justify content to end", Example: `<div class="flex justify-end">`},
		{Name: "justify-between", Category: "flex", Description: "Space between items", Example: `<div class="flex justify-between">`},
		{Name: "justify-around", Category: "flex", Description: "Space around items", Example: `<div class="flex justify-around">`},
		{Name: "justify-evenly", Category: "flex", Description: "Space evenly between items", Example: `<div class="flex justify-evenly">`},
	}
	classes = append(classes, justifyClasses...)

	// Align items
	alignClasses := []TailwindClassInfo{
		{Name: "items-start", Category: "flex", Description: "Align items to start", Example: `<div class="flex items-start">`},
		{Name: "items-center", Category: "flex", Description: "Align items to center", Example: `<div class="flex items-center">`},
		{Name: "items-end", Category: "flex", Description: "Align items to end", Example: `<div class="flex items-end">`},
		{Name: "items-stretch", Category: "flex", Description: "Stretch items to fill", Example: `<div class="flex items-stretch">`},
	}
	classes = append(classes, alignClasses...)

	// Self alignment
	selfClasses := []TailwindClassInfo{
		{Name: "self-start", Category: "flex", Description: "Align self to start", Example: `<div class="self-start">`},
		{Name: "self-center", Category: "flex", Description: "Align self to center", Example: `<div class="self-center">`},
		{Name: "self-end", Category: "flex", Description: "Align self to end", Example: `<div class="self-end">`},
		{Name: "self-stretch", Category: "flex", Description: "Stretch self to fill", Example: `<div class="self-stretch">`},
	}
	classes = append(classes, selfClasses...)

	// Gap classes
	gapClasses := []TailwindClassInfo{
		{Name: "gap-1", Category: "spacing", Description: "Gap of 1 character", Example: `<div class="flex gap-1">`},
		{Name: "gap-2", Category: "spacing", Description: "Gap of 2 characters", Example: `<div class="flex gap-2">`},
		{Name: "gap-3", Category: "spacing", Description: "Gap of 3 characters", Example: `<div class="flex gap-3">`},
		{Name: "gap-4", Category: "spacing", Description: "Gap of 4 characters", Example: `<div class="flex gap-4">`},
	}
	classes = append(classes, gapClasses...)

	// Padding classes
	paddingClasses := []TailwindClassInfo{
		{Name: "p-1", Category: "spacing", Description: "Padding of 1 on all sides", Example: `<div class="p-1">`},
		{Name: "p-2", Category: "spacing", Description: "Padding of 2 on all sides", Example: `<div class="p-2">`},
		{Name: "p-3", Category: "spacing", Description: "Padding of 3 on all sides", Example: `<div class="p-3">`},
		{Name: "p-4", Category: "spacing", Description: "Padding of 4 on all sides", Example: `<div class="p-4">`},
		{Name: "px-1", Category: "spacing", Description: "Horizontal padding of 1", Example: `<div class="px-1">`},
		{Name: "px-2", Category: "spacing", Description: "Horizontal padding of 2", Example: `<div class="px-2">`},
		{Name: "py-1", Category: "spacing", Description: "Vertical padding of 1", Example: `<div class="py-1">`},
		{Name: "py-2", Category: "spacing", Description: "Vertical padding of 2", Example: `<div class="py-2">`},
		{Name: "pt-1", Category: "spacing", Description: "Top padding of 1", Example: `<div class="pt-1">`},
		{Name: "pt-2", Category: "spacing", Description: "Top padding of 2", Example: `<div class="pt-2">`},
		{Name: "pr-1", Category: "spacing", Description: "Right padding of 1", Example: `<div class="pr-1">`},
		{Name: "pr-2", Category: "spacing", Description: "Right padding of 2", Example: `<div class="pr-2">`},
		{Name: "pb-1", Category: "spacing", Description: "Bottom padding of 1", Example: `<div class="pb-1">`},
		{Name: "pb-2", Category: "spacing", Description: "Bottom padding of 2", Example: `<div class="pb-2">`},
		{Name: "pl-1", Category: "spacing", Description: "Left padding of 1", Example: `<div class="pl-1">`},
		{Name: "pl-2", Category: "spacing", Description: "Left padding of 2", Example: `<div class="pl-2">`},
	}
	classes = append(classes, paddingClasses...)

	// Margin classes
	marginClasses := []TailwindClassInfo{
		{Name: "m-1", Category: "spacing", Description: "Margin of 1 on all sides", Example: `<div class="m-1">`},
		{Name: "m-2", Category: "spacing", Description: "Margin of 2 on all sides", Example: `<div class="m-2">`},
		{Name: "m-3", Category: "spacing", Description: "Margin of 3 on all sides", Example: `<div class="m-3">`},
		{Name: "m-4", Category: "spacing", Description: "Margin of 4 on all sides", Example: `<div class="m-4">`},
		{Name: "mx-1", Category: "spacing", Description: "Horizontal margin of 1", Example: `<div class="mx-1">`},
		{Name: "mx-2", Category: "spacing", Description: "Horizontal margin of 2", Example: `<div class="mx-2">`},
		{Name: "my-1", Category: "spacing", Description: "Vertical margin of 1", Example: `<div class="my-1">`},
		{Name: "my-2", Category: "spacing", Description: "Vertical margin of 2", Example: `<div class="my-2">`},
		{Name: "mt-1", Category: "spacing", Description: "Top margin of 1", Example: `<div class="mt-1">`},
		{Name: "mt-2", Category: "spacing", Description: "Top margin of 2", Example: `<div class="mt-2">`},
		{Name: "mr-1", Category: "spacing", Description: "Right margin of 1", Example: `<div class="mr-1">`},
		{Name: "mr-2", Category: "spacing", Description: "Right margin of 2", Example: `<div class="mr-2">`},
		{Name: "mb-1", Category: "spacing", Description: "Bottom margin of 1", Example: `<div class="mb-1">`},
		{Name: "mb-2", Category: "spacing", Description: "Bottom margin of 2", Example: `<div class="mb-2">`},
		{Name: "ml-1", Category: "spacing", Description: "Left margin of 1", Example: `<div class="ml-1">`},
		{Name: "ml-2", Category: "spacing", Description: "Left margin of 2", Example: `<div class="ml-2">`},
	}
	classes = append(classes, marginClasses...)

	// Width classes
	widthClasses := []TailwindClassInfo{
		{Name: "w-full", Category: "layout", Description: "Full width (100%)", Example: `<div class="w-full">`},
		{Name: "w-auto", Category: "layout", Description: "Auto width (size to content)", Example: `<div class="w-auto">`},
		{Name: "w-1/2", Category: "layout", Description: "Half width (50%)", Example: `<div class="w-1/2">`},
		{Name: "w-1/3", Category: "layout", Description: "One-third width (33%)", Example: `<div class="w-1/3">`},
		{Name: "w-2/3", Category: "layout", Description: "Two-thirds width (67%)", Example: `<div class="w-2/3">`},
		{Name: "w-1/4", Category: "layout", Description: "Quarter width (25%)", Example: `<div class="w-1/4">`},
		{Name: "w-3/4", Category: "layout", Description: "Three-quarters width (75%)", Example: `<div class="w-3/4">`},
	}
	classes = append(classes, widthClasses...)

	// Height classes
	heightClasses := []TailwindClassInfo{
		{Name: "h-full", Category: "layout", Description: "Full height (100%)", Example: `<div class="h-full">`},
		{Name: "h-auto", Category: "layout", Description: "Auto height (size to content)", Example: `<div class="h-auto">`},
		{Name: "h-1/2", Category: "layout", Description: "Half height (50%)", Example: `<div class="h-1/2">`},
		{Name: "h-1/3", Category: "layout", Description: "One-third height (33%)", Example: `<div class="h-1/3">`},
		{Name: "h-2/3", Category: "layout", Description: "Two-thirds height (67%)", Example: `<div class="h-2/3">`},
		{Name: "h-1/4", Category: "layout", Description: "Quarter height (25%)", Example: `<div class="h-1/4">`},
		{Name: "h-3/4", Category: "layout", Description: "Three-quarters height (75%)", Example: `<div class="h-3/4">`},
	}
	classes = append(classes, heightClasses...)

	// Border classes
	borderClasses := []TailwindClassInfo{
		{Name: "border", Category: "visual", Description: "Single line border", Example: `<div class="border">`},
		{Name: "border-rounded", Category: "visual", Description: "Rounded border", Example: `<div class="border-rounded">`},
		{Name: "border-double", Category: "visual", Description: "Double line border", Example: `<div class="border-double">`},
		{Name: "border-thick", Category: "visual", Description: "Thick border", Example: `<div class="border-thick">`},
		{Name: "border-red", Category: "visual", Description: "Red border color", Example: `<div class="border border-red">`},
		{Name: "border-green", Category: "visual", Description: "Green border color", Example: `<div class="border border-green">`},
		{Name: "border-blue", Category: "visual", Description: "Blue border color", Example: `<div class="border border-blue">`},
		{Name: "border-cyan", Category: "visual", Description: "Cyan border color", Example: `<div class="border border-cyan">`},
		{Name: "border-magenta", Category: "visual", Description: "Magenta border color", Example: `<div class="border border-magenta">`},
		{Name: "border-yellow", Category: "visual", Description: "Yellow border color", Example: `<div class="border border-yellow">`},
		{Name: "border-white", Category: "visual", Description: "White border color", Example: `<div class="border border-white">`},
		{Name: "border-black", Category: "visual", Description: "Black border color", Example: `<div class="border border-black">`},
	}
	classes = append(classes, borderClasses...)

	// Typography classes
	typographyClasses := []TailwindClassInfo{
		{Name: "font-bold", Category: "typography", Description: "Bold text", Example: `<span class="font-bold">Bold</span>`},
		{Name: "font-dim", Category: "typography", Description: "Dim/faint text", Example: `<span class="font-dim">Dim</span>`},
		{Name: "italic", Category: "typography", Description: "Italic text", Example: `<span class="italic">Italic</span>`},
		{Name: "underline", Category: "typography", Description: "Underlined text", Example: `<span class="underline">Underlined</span>`},
		{Name: "strikethrough", Category: "typography", Description: "Strikethrough text", Example: `<span class="strikethrough">Strikethrough</span>`},
		{Name: "text-left", Category: "typography", Description: "Align text left", Example: `<div class="text-left">`},
		{Name: "text-center", Category: "typography", Description: "Center text", Example: `<div class="text-center">`},
		{Name: "text-right", Category: "typography", Description: "Align text right", Example: `<div class="text-right">`},
	}
	classes = append(classes, typographyClasses...)

	// Text color classes
	textColorClasses := []TailwindClassInfo{
		{Name: "text-red", Category: "visual", Description: "Red text color", Example: `<span class="text-red">Red</span>`},
		{Name: "text-green", Category: "visual", Description: "Green text color", Example: `<span class="text-green">Green</span>`},
		{Name: "text-blue", Category: "visual", Description: "Blue text color", Example: `<span class="text-blue">Blue</span>`},
		{Name: "text-cyan", Category: "visual", Description: "Cyan text color", Example: `<span class="text-cyan">Cyan</span>`},
		{Name: "text-magenta", Category: "visual", Description: "Magenta text color", Example: `<span class="text-magenta">Magenta</span>`},
		{Name: "text-yellow", Category: "visual", Description: "Yellow text color", Example: `<span class="text-yellow">Yellow</span>`},
		{Name: "text-white", Category: "visual", Description: "White text color", Example: `<span class="text-white">White</span>`},
		{Name: "text-black", Category: "visual", Description: "Black text color", Example: `<span class="text-black">Black</span>`},
	}
	classes = append(classes, textColorClasses...)

	// Background color classes
	bgColorClasses := []TailwindClassInfo{
		{Name: "bg-red", Category: "visual", Description: "Red background", Example: `<div class="bg-red">`},
		{Name: "bg-green", Category: "visual", Description: "Green background", Example: `<div class="bg-green">`},
		{Name: "bg-blue", Category: "visual", Description: "Blue background", Example: `<div class="bg-blue">`},
		{Name: "bg-cyan", Category: "visual", Description: "Cyan background", Example: `<div class="bg-cyan">`},
		{Name: "bg-magenta", Category: "visual", Description: "Magenta background", Example: `<div class="bg-magenta">`},
		{Name: "bg-yellow", Category: "visual", Description: "Yellow background", Example: `<div class="bg-yellow">`},
		{Name: "bg-white", Category: "visual", Description: "White background", Example: `<div class="bg-white">`},
		{Name: "bg-black", Category: "visual", Description: "Black background", Example: `<div class="bg-black">`},
	}
	classes = append(classes, bgColorClasses...)

	// Scroll classes
	scrollClasses := []TailwindClassInfo{
		{Name: "overflow-scroll", Category: "layout", Description: "Enable scrolling in both directions", Example: `<div class="overflow-scroll">`},
		{Name: "overflow-y-scroll", Category: "layout", Description: "Enable vertical scrolling", Example: `<div class="overflow-y-scroll">`},
		{Name: "overflow-x-scroll", Category: "layout", Description: "Enable horizontal scrolling", Example: `<div class="overflow-x-scroll">`},
	}
	classes = append(classes, scrollClasses...)

	return classes
}
