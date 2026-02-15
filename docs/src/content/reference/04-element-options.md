---
title: "Element Options"
order: 4
---

Functional options for New().

## WithText

```go
func WithText(text string) Option
```

Sets the text content.

## WithSize

```go
func WithSize(w, h int) Option
```

Sets fixed width and height.

## WithDirection

```go
func WithDirection(d Direction) Option
```

Sets flex direction (Row or Column).

## WithFlexGrow

```go
func WithFlexGrow(v float64) Option
```

Sets flex grow factor.

## WithFlexShrink

```go
func WithFlexShrink(v float64) Option
```

Sets flex shrink factor.

## WithJustify

```go
func WithJustify(j Justify) Option
```

Sets main-axis alignment.

## WithAlign

```go
func WithAlign(a Align) Option
```

Sets cross-axis alignment.

## WithGap

```go
func WithGap(g int) Option
```

Sets gap between children.

## WithPadding

```go
func WithPadding(p int) Option
```

Sets padding on all sides.

## WithMargin

```go
func WithMargin(m int) Option
```

Sets margin on all sides.

## WithBorder

```go
func WithBorder(b BorderStyle) Option
```

Sets border style (BorderSingle, BorderDouble, BorderRounded, BorderThick).

## WithTextStyle

```go
func WithTextStyle(s Style) Option
```

Sets the text rendering style.

## WithBackground

```go
func WithBackground(s Style) Option
```

Sets the background style.

## WithOnFocus

```go
func WithOnFocus(fn func(*Element)) Option
```

Sets focus callback.

## WithOnBlur

```go
func WithOnBlur(fn func(*Element)) Option
```

Sets blur callback.
