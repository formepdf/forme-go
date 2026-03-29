// Package templates provides a component DSL for building Forme PDF documents.
//
// Components serialize to the JSON schema that the Rust engine expects.
// WASM is only needed at render time (DocumentNode.Render), not for building
// or serializing the document tree.
package templates

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Node is the interface implemented by all Forme components.
type Node interface {
	toDict() map[string]any
}

// ── Style ───────────────────────────────────────────────────────────

// Style configures the visual properties of a component.
// Use `any` for Width/Height to support float64, "auto", "50%".
type Style struct {
	// Dimensions
	Width     any `json:"width,omitempty"`
	Height    any `json:"height,omitempty"`
	MinWidth  any `json:"minWidth,omitempty"`
	MinHeight any `json:"minHeight,omitempty"`
	MaxWidth  any `json:"maxWidth,omitempty"`
	MaxHeight any `json:"maxHeight,omitempty"`

	// Padding (shorthand: uniform value; individual overrides)
	Padding           any     `json:"padding,omitempty"`
	PaddingTop        float64 `json:"paddingTop,omitempty"`
	PaddingRight      float64 `json:"paddingRight,omitempty"`
	PaddingBottom     float64 `json:"paddingBottom,omitempty"`
	PaddingLeft       float64 `json:"paddingLeft,omitempty"`
	PaddingHorizontal float64 `json:"paddingHorizontal,omitempty"`
	PaddingVertical   float64 `json:"paddingVertical,omitempty"`

	// Margin (shorthand: uniform value; individual overrides)
	Margin           any     `json:"margin,omitempty"`
	MarginTop        float64 `json:"marginTop,omitempty"`
	MarginRight      float64 `json:"marginRight,omitempty"`
	MarginBottom     float64 `json:"marginBottom,omitempty"`
	MarginLeft       float64 `json:"marginLeft,omitempty"`
	MarginHorizontal float64 `json:"marginHorizontal,omitempty"`
	MarginVertical   float64 `json:"marginVertical,omitempty"`

	// Flex
	Flex           float64 `json:"flex,omitempty"`
	FlexDirection  string  `json:"flexDirection,omitempty"`
	FlexGrow       float64 `json:"flexGrow,omitempty"`
	FlexShrink     float64 `json:"flexShrink,omitempty"`
	FlexBasis      any     `json:"flexBasis,omitempty"`
	FlexWrap       string  `json:"flexWrap,omitempty"`
	JustifyContent string  `json:"justifyContent,omitempty"`
	AlignItems     string  `json:"alignItems,omitempty"`
	AlignSelf      string  `json:"alignSelf,omitempty"`
	AlignContent   string  `json:"alignContent,omitempty"`
	Gap            float64 `json:"gap,omitempty"`
	RowGap         float64 `json:"rowGap,omitempty"`
	ColumnGap      float64 `json:"columnGap,omitempty"`

	// Display
	Display string `json:"display,omitempty"`

	// Grid
	GridTemplateColumns string `json:"gridTemplateColumns,omitempty"`
	GridTemplateRows    string `json:"gridTemplateRows,omitempty"`
	GridAutoRows        string `json:"gridAutoRows,omitempty"`
	GridAutoColumns     string `json:"gridAutoColumns,omitempty"`
	GridColumnStart     int    `json:"gridColumnStart,omitempty"`
	GridColumnEnd       int    `json:"gridColumnEnd,omitempty"`
	GridRowStart        int    `json:"gridRowStart,omitempty"`
	GridRowEnd          int    `json:"gridRowEnd,omitempty"`
	GridColumnSpan      int    `json:"gridColumnSpan,omitempty"`
	GridRowSpan         int    `json:"gridRowSpan,omitempty"`

	// Typography
	FontFamily     string  `json:"fontFamily,omitempty"`
	FontSize       float64 `json:"fontSize,omitempty"`
	FontWeight     any     `json:"fontWeight,omitempty"` // int, "bold", "normal"
	FontStyle      string  `json:"fontStyle,omitempty"`
	LineHeight     float64 `json:"lineHeight,omitempty"`
	TextAlign      string  `json:"textAlign,omitempty"`
	LetterSpacing  float64 `json:"letterSpacing,omitempty"`
	TextDecoration string  `json:"textDecoration,omitempty"`
	TextTransform  string  `json:"textTransform,omitempty"`
	Hyphens        string  `json:"hyphens,omitempty"`
	Lang           string  `json:"lang,omitempty"`
	Direction      string  `json:"direction,omitempty"`
	TextOverflow   string  `json:"textOverflow,omitempty"`
	LineBreaking   string  `json:"lineBreaking,omitempty"`
	Overflow       string  `json:"overflow,omitempty"`

	// Color
	Color           string  `json:"color,omitempty"`
	BackgroundColor string  `json:"backgroundColor,omitempty"`
	Opacity         float64 `json:"opacity,omitempty"`

	// Border
	Border             string  `json:"border,omitempty"`
	BorderTop          any     `json:"borderTop,omitempty"`
	BorderRight        any     `json:"borderRight,omitempty"`
	BorderBottom       any     `json:"borderBottom,omitempty"`
	BorderLeft         any     `json:"borderLeft,omitempty"`
	BorderWidth        any     `json:"borderWidth,omitempty"`
	BorderTopWidth     float64 `json:"borderTopWidth,omitempty"`
	BorderRightWidth   float64 `json:"borderRightWidth,omitempty"`
	BorderBottomWidth  float64 `json:"borderBottomWidth,omitempty"`
	BorderLeftWidth    float64 `json:"borderLeftWidth,omitempty"`
	BorderColor        any     `json:"borderColor,omitempty"` // string or map[string]string
	BorderTopColor     string  `json:"borderTopColor,omitempty"`
	BorderRightColor   string  `json:"borderRightColor,omitempty"`
	BorderBottomColor  string  `json:"borderBottomColor,omitempty"`
	BorderLeftColor    string  `json:"borderLeftColor,omitempty"`
	BorderRadius       float64 `json:"borderRadius,omitempty"`
	BorderTopLeftRadius     float64 `json:"borderTopLeftRadius,omitempty"`
	BorderTopRightRadius    float64 `json:"borderTopRightRadius,omitempty"`
	BorderBottomRightRadius float64 `json:"borderBottomRightRadius,omitempty"`
	BorderBottomLeftRadius  float64 `json:"borderBottomLeftRadius,omitempty"`

	// Positioning
	Position string  `json:"position,omitempty"`
	Top      float64 `json:"top,omitempty"`
	Right    float64 `json:"right,omitempty"`
	Bottom   float64 `json:"bottom,omitempty"`
	Left     float64 `json:"left,omitempty"`

	// Page behavior
	Wrap           *bool `json:"wrap,omitempty"`
	BreakBefore    *bool `json:"breakBefore,omitempty"`
	MinWidowLines  int   `json:"minWidowLines,omitempty"`
	MinOrphanLines int   `json:"minOrphanLines,omitempty"`
}

// ── Style mapping ───────────────────────────────────────────────────

var flexDirectionMap = map[string]string{
	"row": "Row", "column": "Column",
	"row-reverse": "RowReverse", "column-reverse": "ColumnReverse",
}

var justifyContentMap = map[string]string{
	"flex-start": "FlexStart", "flex-end": "FlexEnd", "center": "Center",
	"space-between": "SpaceBetween", "space-around": "SpaceAround", "space-evenly": "SpaceEvenly",
}

var alignItemsMap = map[string]string{
	"flex-start": "FlexStart", "flex-end": "FlexEnd", "center": "Center",
	"stretch": "Stretch", "baseline": "Baseline",
}

var flexWrapMap = map[string]string{
	"nowrap": "NoWrap", "wrap": "Wrap",
}

var alignContentMap = map[string]string{
	"flex-start": "FlexStart", "flex-end": "FlexEnd", "center": "Center",
	"space-between": "SpaceBetween", "space-around": "SpaceAround", "space-evenly": "SpaceEvenly",
	"stretch": "Stretch",
}

var fontStyleMap = map[string]string{
	"normal": "Normal", "italic": "Italic", "oblique": "Oblique",
}

var textAlignMap = map[string]string{
	"left": "Left", "right": "Right", "center": "Center", "justify": "Justify",
}

var textDecorationMap = map[string]string{
	"none": "None", "underline": "Underline", "line-through": "LineThrough",
}

var textTransformMap = map[string]string{
	"none": "None", "uppercase": "Uppercase", "lowercase": "Lowercase", "capitalize": "Capitalize",
}

var hyphensMap = map[string]string{
	"none": "none", "manual": "manual", "auto": "auto",
}

var textOverflowMap = map[string]string{
	"wrap": "Wrap", "ellipsis": "Ellipsis", "clip": "Clip",
}

var lineBreakingMap = map[string]string{
	"optimal": "optimal", "greedy": "greedy",
}

var overflowMap = map[string]string{
	"visible": "Visible", "hidden": "Hidden",
}

var displayMap = map[string]string{
	"flex": "Flex", "grid": "Grid",
}

var positionMap = map[string]string{
	"relative": "Relative", "absolute": "Absolute",
}

var percentRe = regexp.MustCompile(`^([0-9.]+)%$`)
var frRe = regexp.MustCompile(`^([0-9.]+)fr$`)
var rgbaRe = regexp.MustCompile(`^rgba\(\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)\s*\)$`)
var rgbRe = regexp.MustCompile(`^rgb\(\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)\s*\)$`)
var repeatRe = regexp.MustCompile(`repeat\(\s*(\d+)\s*,\s*([^)]+)\)`)

func parseColor(color string) map[string]float64 {
	s := strings.TrimSpace(color)

	if m := rgbaRe.FindStringSubmatch(s); m != nil {
		r, _ := strconv.ParseFloat(m[1], 64)
		g, _ := strconv.ParseFloat(m[2], 64)
		b, _ := strconv.ParseFloat(m[3], 64)
		a, _ := strconv.ParseFloat(m[4], 64)
		return map[string]float64{"r": r / 255, "g": g / 255, "b": b / 255, "a": a}
	}

	if m := rgbRe.FindStringSubmatch(s); m != nil {
		r, _ := strconv.ParseFloat(m[1], 64)
		g, _ := strconv.ParseFloat(m[2], 64)
		b, _ := strconv.ParseFloat(m[3], 64)
		return map[string]float64{"r": r / 255, "g": g / 255, "b": b / 255, "a": 1.0}
	}

	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) == 3 {
			h = string(h[0]) + string(h[0]) + string(h[1]) + string(h[1]) + string(h[2]) + string(h[2])
		}
		if len(h) == 6 {
			r, _ := strconv.ParseInt(h[0:2], 16, 64)
			g, _ := strconv.ParseInt(h[2:4], 16, 64)
			b, _ := strconv.ParseInt(h[4:6], 16, 64)
			return map[string]float64{"r": float64(r) / 255, "g": float64(g) / 255, "b": float64(b) / 255, "a": 1.0}
		}
		if len(h) == 8 {
			r, _ := strconv.ParseInt(h[0:2], 16, 64)
			g, _ := strconv.ParseInt(h[2:4], 16, 64)
			b, _ := strconv.ParseInt(h[4:6], 16, 64)
			a, _ := strconv.ParseInt(h[6:8], 16, 64)
			return map[string]float64{"r": float64(r) / 255, "g": float64(g) / 255, "b": float64(b) / 255, "a": float64(a) / 255}
		}
	}

	return map[string]float64{"r": 0, "g": 0, "b": 0, "a": 1}
}

func mapDimension(val any) any {
	switch v := val.(type) {
	case int:
		return map[string]any{"Pt": float64(v)}
	case float64:
		return map[string]any{"Pt": v}
	case string:
		if v == "auto" {
			return "Auto"
		}
		if m := percentRe.FindStringSubmatch(v); m != nil {
			pct, _ := strconv.ParseFloat(m[1], 64)
			return map[string]any{"Percent": pct}
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return map[string]any{"Pt": f}
		}
		return "Auto"
	default:
		return "Auto"
	}
}

func expandEdges(val any) map[string]float64 {
	switch v := val.(type) {
	case int:
		f := float64(v)
		return map[string]float64{"top": f, "right": f, "bottom": f, "left": f}
	case float64:
		return map[string]float64{"top": v, "right": v, "bottom": v, "left": v}
	case []float64:
		switch len(v) {
		case 1:
			return map[string]float64{"top": v[0], "right": v[0], "bottom": v[0], "left": v[0]}
		case 2:
			return map[string]float64{"top": v[0], "right": v[1], "bottom": v[0], "left": v[1]}
		case 3:
			return map[string]float64{"top": v[0], "right": v[1], "bottom": v[2], "left": v[1]}
		default:
			return map[string]float64{"top": v[0], "right": v[1], "bottom": v[2], "left": v[3]}
		}
	default:
		return map[string]float64{"top": 0, "right": 0, "bottom": 0, "left": 0}
	}
}

func expandCorners(val float64) map[string]float64 {
	return map[string]float64{
		"top_left": val, "top_right": val, "bottom_right": val, "bottom_left": val,
	}
}

var borderStyleKeywords = map[string]bool{
	"solid": true, "dashed": true, "dotted": true, "double": true,
	"groove": true, "ridge": true, "inset": true, "outset": true,
	"none": true, "hidden": true,
}

func parseBorderString(val string) (width *float64, color map[string]float64) {
	tokens := strings.Fields(val)
	for _, token := range tokens {
		lower := strings.ToLower(token)
		if borderStyleKeywords[lower] {
			continue
		}
		if strings.HasPrefix(token, "#") || strings.HasPrefix(token, "rgb") {
			color = parseColor(token)
			continue
		}
		cleaned := strings.TrimSuffix(token, "px")
		if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
			width = &f
		}
	}
	return
}

func mapGridTrack(track string) any {
	if track == "auto" {
		return "Auto"
	}
	if m := frRe.FindStringSubmatch(track); m != nil {
		f, _ := strconv.ParseFloat(m[1], 64)
		return map[string]any{"Fr": f}
	}
	if f, err := strconv.ParseFloat(track, 64); err == nil {
		return map[string]any{"Pt": f}
	}
	return "Auto"
}

func expandRepeat(s string) string {
	return repeatRe.ReplaceAllStringFunc(s, func(match string) string {
		m := repeatRe.FindStringSubmatch(match)
		count, _ := strconv.Atoi(m[1])
		tracks := strings.TrimSpace(m[2])
		parts := make([]string, count)
		for i := range parts {
			parts[i] = tracks
		}
		return strings.Join(parts, " ")
	})
}

func parseGridTemplate(value string) []any {
	expanded := expandRepeat(value)
	tokens := strings.Fields(expanded)
	result := make([]any, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, mapGridTrack(t))
	}
	return result
}

func mapStyle(s *Style) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	result := map[string]any{}

	// Dimensions
	for _, pair := range []struct {
		key string
		val any
	}{
		{"width", s.Width}, {"height", s.Height},
		{"minWidth", s.MinWidth}, {"minHeight", s.MinHeight},
		{"maxWidth", s.MaxWidth}, {"maxHeight", s.MaxHeight},
	} {
		if pair.val != nil {
			result[pair.key] = mapDimension(pair.val)
		}
	}

	// Padding
	hasPad := s.Padding != nil || s.PaddingTop != 0 || s.PaddingRight != 0 ||
		s.PaddingBottom != 0 || s.PaddingLeft != 0 || s.PaddingHorizontal != 0 || s.PaddingVertical != 0
	if hasPad {
		base := map[string]float64{"top": 0, "right": 0, "bottom": 0, "left": 0}
		if s.Padding != nil {
			base = expandEdges(s.Padding)
		}
		top, right, bottom, left := base["top"], base["right"], base["bottom"], base["left"]
		if s.PaddingVertical != 0 {
			top = s.PaddingVertical
			bottom = s.PaddingVertical
		}
		if s.PaddingHorizontal != 0 {
			left = s.PaddingHorizontal
			right = s.PaddingHorizontal
		}
		if s.PaddingTop != 0 {
			top = s.PaddingTop
		}
		if s.PaddingRight != 0 {
			right = s.PaddingRight
		}
		if s.PaddingBottom != 0 {
			bottom = s.PaddingBottom
		}
		if s.PaddingLeft != 0 {
			left = s.PaddingLeft
		}
		result["padding"] = map[string]float64{"top": top, "right": right, "bottom": bottom, "left": left}
	}

	// Margin
	hasMargin := s.Margin != nil || s.MarginTop != 0 || s.MarginRight != 0 ||
		s.MarginBottom != 0 || s.MarginLeft != 0 || s.MarginHorizontal != 0 || s.MarginVertical != 0
	if hasMargin {
		base := map[string]float64{"top": 0, "right": 0, "bottom": 0, "left": 0}
		if s.Margin != nil {
			base = expandEdges(s.Margin)
		}
		top, right, bottom, left := base["top"], base["right"], base["bottom"], base["left"]
		if s.MarginVertical != 0 {
			top = s.MarginVertical
			bottom = s.MarginVertical
		}
		if s.MarginHorizontal != 0 {
			left = s.MarginHorizontal
			right = s.MarginHorizontal
		}
		if s.MarginTop != 0 {
			top = s.MarginTop
		}
		if s.MarginRight != 0 {
			right = s.MarginRight
		}
		if s.MarginBottom != 0 {
			bottom = s.MarginBottom
		}
		if s.MarginLeft != 0 {
			left = s.MarginLeft
		}
		result["margin"] = map[string]float64{"top": top, "right": right, "bottom": bottom, "left": left}
	}

	// Flex shorthand
	if s.Flex != 0 {
		if s.FlexGrow == 0 {
			result["flexGrow"] = s.Flex
		}
		if s.FlexShrink == 0 {
			result["flexShrink"] = float64(1)
		}
		if s.FlexBasis == nil {
			result["flexBasis"] = map[string]any{"Pt": float64(0)}
		}
	}

	// Flex enums
	if s.FlexDirection != "" {
		result["flexDirection"] = mapEnum(flexDirectionMap, s.FlexDirection)
	}
	if s.JustifyContent != "" {
		result["justifyContent"] = mapEnum(justifyContentMap, s.JustifyContent)
	}
	if s.AlignItems != "" {
		result["alignItems"] = mapEnum(alignItemsMap, s.AlignItems)
	}
	if s.AlignSelf != "" {
		result["alignSelf"] = mapEnum(alignItemsMap, s.AlignSelf)
	}
	if s.FlexWrap != "" {
		result["flexWrap"] = mapEnum(flexWrapMap, s.FlexWrap)
	}
	if s.AlignContent != "" {
		result["alignContent"] = mapEnum(alignContentMap, s.AlignContent)
	}

	// Flex numeric
	if s.FlexGrow != 0 {
		result["flexGrow"] = s.FlexGrow
	}
	if s.FlexShrink != 0 {
		result["flexShrink"] = s.FlexShrink
	}
	if s.Gap != 0 {
		result["gap"] = s.Gap
	}
	if s.RowGap != 0 {
		result["rowGap"] = s.RowGap
	}
	if s.ColumnGap != 0 {
		result["columnGap"] = s.ColumnGap
	}
	if s.FlexBasis != nil {
		result["flexBasis"] = mapDimension(s.FlexBasis)
	}

	// Display
	if s.Display != "" {
		if v, ok := displayMap[s.Display]; ok {
			result["display"] = v
		} else {
			result["display"] = "Flex"
		}
	}

	// Grid
	if s.GridTemplateColumns != "" {
		result["gridTemplateColumns"] = parseGridTemplate(s.GridTemplateColumns)
	}
	if s.GridTemplateRows != "" {
		result["gridTemplateRows"] = parseGridTemplate(s.GridTemplateRows)
	}
	if s.GridAutoRows != "" {
		result["gridAutoRows"] = mapGridTrack(s.GridAutoRows)
	}
	if s.GridAutoColumns != "" {
		result["gridAutoColumns"] = mapGridTrack(s.GridAutoColumns)
	}
	hasGridPlacement := s.GridColumnStart != 0 || s.GridColumnEnd != 0 ||
		s.GridRowStart != 0 || s.GridRowEnd != 0 || s.GridColumnSpan != 0 || s.GridRowSpan != 0
	if hasGridPlacement {
		placement := map[string]any{}
		if s.GridColumnStart != 0 {
			placement["columnStart"] = s.GridColumnStart
		}
		if s.GridColumnEnd != 0 {
			placement["columnEnd"] = s.GridColumnEnd
		}
		if s.GridRowStart != 0 {
			placement["rowStart"] = s.GridRowStart
		}
		if s.GridRowEnd != 0 {
			placement["rowEnd"] = s.GridRowEnd
		}
		if s.GridColumnSpan != 0 {
			placement["columnSpan"] = s.GridColumnSpan
		}
		if s.GridRowSpan != 0 {
			placement["rowSpan"] = s.GridRowSpan
		}
		result["gridPlacement"] = placement
	}

	// Typography
	if s.FontFamily != "" {
		result["fontFamily"] = s.FontFamily
	}
	if s.FontSize != 0 {
		result["fontSize"] = s.FontSize
	}
	if s.FontWeight != nil {
		switch fw := s.FontWeight.(type) {
		case string:
			if fw == "bold" {
				result["fontWeight"] = 700
			} else if fw == "normal" {
				result["fontWeight"] = 400
			} else {
				result["fontWeight"] = fw
			}
		case int:
			result["fontWeight"] = fw
		case float64:
			result["fontWeight"] = int(fw)
		}
	}
	if s.FontStyle != "" {
		result["fontStyle"] = mapEnum(fontStyleMap, s.FontStyle)
	}
	if s.LineHeight != 0 {
		result["lineHeight"] = s.LineHeight
	}
	if s.TextAlign != "" {
		result["textAlign"] = mapEnum(textAlignMap, s.TextAlign)
	}
	if s.LetterSpacing != 0 {
		result["letterSpacing"] = s.LetterSpacing
	}
	if s.TextDecoration != "" {
		result["textDecoration"] = mapEnum(textDecorationMap, s.TextDecoration)
	}
	if s.TextTransform != "" {
		result["textTransform"] = mapEnum(textTransformMap, s.TextTransform)
	}
	if s.Hyphens != "" {
		result["hyphens"] = mapEnum(hyphensMap, s.Hyphens)
	}
	if s.Lang != "" {
		result["lang"] = s.Lang
	}
	if s.Direction != "" {
		result["direction"] = s.Direction
	}
	if s.TextOverflow != "" {
		result["textOverflow"] = mapEnum(textOverflowMap, s.TextOverflow)
	}
	if s.LineBreaking != "" {
		result["lineBreaking"] = mapEnum(lineBreakingMap, s.LineBreaking)
	}
	if s.Overflow != "" {
		result["overflow"] = mapEnum(overflowMap, s.Overflow)
	}

	// Color
	if s.Color != "" {
		result["color"] = parseColor(s.Color)
	}
	if s.BackgroundColor != "" {
		result["backgroundColor"] = parseColor(s.BackgroundColor)
	}
	if s.Opacity != 0 {
		result["opacity"] = s.Opacity
	}

	// Border
	mapBorder(s, result)

	// Border radius
	hasBR := s.BorderRadius != 0 || s.BorderTopLeftRadius != 0 || s.BorderTopRightRadius != 0 ||
		s.BorderBottomRightRadius != 0 || s.BorderBottomLeftRadius != 0
	if hasBR {
		base := expandCorners(s.BorderRadius)
		if s.BorderTopLeftRadius != 0 {
			base["top_left"] = s.BorderTopLeftRadius
		}
		if s.BorderTopRightRadius != 0 {
			base["top_right"] = s.BorderTopRightRadius
		}
		if s.BorderBottomRightRadius != 0 {
			base["bottom_right"] = s.BorderBottomRightRadius
		}
		if s.BorderBottomLeftRadius != 0 {
			base["bottom_left"] = s.BorderBottomLeftRadius
		}
		result["borderRadius"] = base
	}

	// Positioning
	if s.Position != "" {
		result["position"] = mapEnum(positionMap, s.Position)
	}
	if s.Top != 0 {
		result["top"] = s.Top
	}
	if s.Right != 0 {
		result["right"] = s.Right
	}
	if s.Bottom != 0 {
		result["bottom"] = s.Bottom
	}
	if s.Left != 0 {
		result["left"] = s.Left
	}

	// Page behavior
	if s.Wrap != nil {
		result["wrap"] = *s.Wrap
	}
	if s.BreakBefore != nil {
		result["breakBefore"] = *s.BreakBefore
	}
	if s.MinWidowLines != 0 {
		result["minWidowLines"] = s.MinWidowLines
	}
	if s.MinOrphanLines != 0 {
		result["minOrphanLines"] = s.MinOrphanLines
	}

	return result
}

func mapBorder(s *Style, result map[string]any) {
	type optFloat struct {
		val float64
		set bool
	}
	shortWidth := map[string]optFloat{"top": {}, "right": {}, "bottom": {}, "left": {}}
	shortColor := map[string]map[string]float64{"top": nil, "right": nil, "bottom": nil, "left": nil}

	if s.Border != "" {
		w, c := parseBorderString(s.Border)
		if w != nil {
			for _, side := range []string{"top", "right", "bottom", "left"} {
				shortWidth[side] = optFloat{*w, true}
			}
		}
		if c != nil {
			for _, side := range []string{"top", "right", "bottom", "left"} {
				shortColor[side] = c
			}
		}
	}

	for _, item := range []struct {
		side string
		val  any
	}{
		{"top", s.BorderTop}, {"right", s.BorderRight}, {"bottom", s.BorderBottom}, {"left", s.BorderLeft},
	} {
		if item.val == nil {
			continue
		}
		switch v := item.val.(type) {
		case float64:
			shortWidth[item.side] = optFloat{v, true}
		case int:
			shortWidth[item.side] = optFloat{float64(v), true}
		case string:
			w, c := parseBorderString(v)
			if w != nil {
				shortWidth[item.side] = optFloat{*w, true}
			}
			if c != nil {
				shortColor[item.side] = c
			}
		}
	}

	hasBorderWidth := s.BorderWidth != nil || s.BorderTopWidth != 0 || s.BorderRightWidth != 0 ||
		s.BorderBottomWidth != 0 || s.BorderLeftWidth != 0
	hasShortWidth := false
	for _, v := range shortWidth {
		if v.set {
			hasShortWidth = true
			break
		}
	}
	if hasBorderWidth || hasShortWidth {
		baseBW := map[string]float64{"top": 0, "right": 0, "bottom": 0, "left": 0}
		if s.BorderWidth != nil {
			switch bw := s.BorderWidth.(type) {
			case float64:
				baseBW = map[string]float64{"top": bw, "right": bw, "bottom": bw, "left": bw}
			case int:
				f := float64(bw)
				baseBW = map[string]float64{"top": f, "right": f, "bottom": f, "left": f}
			}
		} else {
			for _, side := range []string{"top", "right", "bottom", "left"} {
				if shortWidth[side].set {
					baseBW[side] = shortWidth[side].val
				}
			}
		}
		bw := map[string]float64{
			"top": baseBW["top"], "right": baseBW["right"],
			"bottom": baseBW["bottom"], "left": baseBW["left"],
		}
		if s.BorderTopWidth != 0 {
			bw["top"] = s.BorderTopWidth
		}
		if s.BorderRightWidth != 0 {
			bw["right"] = s.BorderRightWidth
		}
		if s.BorderBottomWidth != 0 {
			bw["bottom"] = s.BorderBottomWidth
		}
		if s.BorderLeftWidth != 0 {
			bw["left"] = s.BorderLeftWidth
		}
		result["borderWidth"] = bw
	}

	hasBorderColor := s.BorderColor != nil || s.BorderTopColor != "" || s.BorderRightColor != "" ||
		s.BorderBottomColor != "" || s.BorderLeftColor != ""
	hasShortColor := false
	for _, v := range shortColor {
		if v != nil {
			hasShortColor = true
			break
		}
	}
	if hasBorderColor || hasShortColor {
		defaultC := parseColor("#000000")
		baseBC := map[string]map[string]float64{
			"top": defaultC, "right": defaultC, "bottom": defaultC, "left": defaultC,
		}
		for _, side := range []string{"top", "right", "bottom", "left"} {
			if shortColor[side] != nil {
				baseBC[side] = shortColor[side]
			}
		}
		if s.BorderColor != nil {
			switch bc := s.BorderColor.(type) {
			case string:
				c := parseColor(bc)
				for _, side := range []string{"top", "right", "bottom", "left"} {
					baseBC[side] = c
				}
			}
		}
		bc := map[string]any{
			"top": baseBC["top"], "right": baseBC["right"],
			"bottom": baseBC["bottom"], "left": baseBC["left"],
		}
		if s.BorderTopColor != "" {
			bc["top"] = parseColor(s.BorderTopColor)
		}
		if s.BorderRightColor != "" {
			bc["right"] = parseColor(s.BorderRightColor)
		}
		if s.BorderBottomColor != "" {
			bc["bottom"] = parseColor(s.BorderBottomColor)
		}
		if s.BorderLeftColor != "" {
			bc["left"] = parseColor(s.BorderLeftColor)
		}
		result["borderColor"] = bc
	}
}

func mapEnum(m map[string]string, val string) string {
	if mapped, ok := m[val]; ok {
		return mapped
	}
	return val
}

// ── Chart types ─────────────────────────────────────────────────────

// ChartDataPoint is a data point for bar charts and pie charts.
type ChartDataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Color string  `json:"color,omitempty"`
}

// ChartSeries is a data series for line and area charts.
type ChartSeries struct {
	Label  string    `json:"label,omitempty"`
	Data   []float64 `json:"data"`
	Color  string    `json:"color,omitempty"`
}

// DotPlotGroup is a group of dots for a dot plot.
type DotPlotGroup struct {
	Label string       `json:"label,omitempty"`
	Data  []DotPlotPoint `json:"data"`
	Color string       `json:"color,omitempty"`
}

// DotPlotPoint is a single point in a dot plot.
type DotPlotPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Column defines a table column.
type Column struct {
	Width any `json:"width"` // float64, "1fr", "auto"
}

// ── Components ──────────────────────────────────────────────────────

// DocumentNode is the root document node.
type DocumentNode struct {
	children []Node
	title    string
	author   string
	subject  string
	lang     string
	style    *Style
	fonts    []map[string]any
	tagged   bool
}

// Document creates a new document node.
func Document(children ...Node) *DocumentNode {
	return &DocumentNode{children: children}
}

func (d *DocumentNode) Title(title string) *DocumentNode       { d.title = title; return d }
func (d *DocumentNode) Author(author string) *DocumentNode     { d.author = author; return d }
func (d *DocumentNode) Subject(subject string) *DocumentNode   { d.subject = subject; return d }
func (d *DocumentNode) Lang(lang string) *DocumentNode         { d.lang = lang; return d }
func (d *DocumentNode) DefaultStyle(s Style) *DocumentNode     { d.style = &s; return d }
func (d *DocumentNode) Fonts(fonts []map[string]any) *DocumentNode { d.fonts = fonts; return d }
func (d *DocumentNode) Tagged(tagged bool) *DocumentNode       { d.tagged = tagged; return d }

func (d *DocumentNode) toDict() map[string]any {
	doc := map[string]any{
		"children": nodesToDicts(d.children),
	}

	metadata := map[string]any{}
	if d.title != "" {
		metadata["title"] = d.title
	}
	if d.author != "" {
		metadata["author"] = d.author
	}
	if d.subject != "" {
		metadata["subject"] = d.subject
	}
	if d.lang != "" {
		metadata["lang"] = d.lang
	}
	if len(metadata) > 0 {
		doc["metadata"] = metadata
	}

	if d.style != nil {
		doc["default_style"] = mapStyle(d.style)
	}
	if d.fonts != nil {
		doc["fonts"] = d.fonts
	}
	if d.tagged {
		doc["tagged"] = true
	}

	return doc
}

// ToJSON serializes the document to a JSON string.
func (d *DocumentNode) ToJSON() (string, error) {
	data, err := json.Marshal(d.toDict())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Render renders the document to PDF bytes using the local WASM engine.
func (d *DocumentNode) Render(embedData ...any) ([]byte, error) {
	doc := d.toDict()
	if len(embedData) > 0 && embedData[0] != nil {
		data, err := json.Marshal(embedData[0])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embed data: %w", err)
		}
		doc["embedded_data"] = string(data)
	}
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return renderPDF(string(jsonBytes))
}

// PageNode represents a page.
type PageNode struct {
	children []Node
	size     any // string or map
	margin   any
}

// Page creates a new page node.
func Page(children ...Node) *PageNode {
	return &PageNode{children: children}
}

func (p *PageNode) Size(size any) *PageNode     { p.size = size; return p }
func (p *PageNode) Margin(margin any) *PageNode  { p.margin = margin; return p }

func (p *PageNode) toDict() map[string]any {
	pageSize := any("A4")
	if p.size != nil {
		switch v := p.size.(type) {
		case string:
			pageSize = v
		case map[string]any:
			pageSize = map[string]any{"Custom": map[string]any{"width": v["width"], "height": v["height"]}}
		}
	}

	pageMargin := map[string]float64{"top": 54, "right": 54, "bottom": 54, "left": 54}
	if p.margin != nil {
		pageMargin = expandEdges(p.margin)
	}

	return map[string]any{
		"kind": map[string]any{
			"type": "Page",
			"config": map[string]any{
				"size":   pageSize,
				"margin": pageMargin,
				"wrap":   true,
			},
		},
		"style":    map[string]any{},
		"children": nodesToDicts(p.children),
	}
}

// ViewNode represents a flex/grid container.
type ViewNode struct {
	children []Node
	style    *Style
	wrap     *bool
	bookmark string
	href     string
}

// View creates a new view node.
func View(children ...Node) *ViewNode {
	return &ViewNode{children: children}
}

func (v *ViewNode) Style(s Style) *ViewNode      { v.style = &s; return v }
func (v *ViewNode) Wrap(wrap bool) *ViewNode      { v.wrap = &wrap; return v }
func (v *ViewNode) Bookmark(b string) *ViewNode   { v.bookmark = b; return v }
func (v *ViewNode) Href(href string) *ViewNode    { v.href = href; return v }

func (v *ViewNode) toDict() map[string]any {
	mapped := mapStyle(v.style)
	if v.wrap != nil {
		mapped["wrap"] = *v.wrap
	}
	node := map[string]any{
		"kind":     map[string]any{"type": "View"},
		"style":    mapped,
		"children": nodesToDicts(v.children),
	}
	if v.bookmark != "" {
		node["bookmark"] = v.bookmark
	}
	if v.href != "" {
		node["href"] = v.href
	}
	return node
}

// TextNode represents a text element.
type TextNode struct {
	content  string
	style    *Style
	href     string
	children []any // string or *TextNode
}

// Text creates a new text node.
func Text(content string, s ...Style) *TextNode {
	t := &TextNode{content: content}
	if len(s) > 0 {
		t.style = &s[0]
	}
	return t
}

func (t *TextNode) Href(href string) *TextNode           { t.href = href; return t }
func (t *TextNode) Style(s Style) *TextNode              { t.style = &s; return t }
func (t *TextNode) Children(children ...any) *TextNode   { t.children = children; return t }

func (t *TextNode) toDict() map[string]any {
	kind := map[string]any{"type": "Text", "content": t.content}

	if len(t.children) > 0 {
		runs := make([]map[string]any, 0, len(t.children))
		for _, child := range t.children {
			switch c := child.(type) {
			case string:
				runs = append(runs, map[string]any{"content": c})
			case *TextNode:
				run := map[string]any{"content": c.content}
				if c.style != nil {
					run["style"] = mapStyle(c.style)
				}
				if c.href != "" {
					run["href"] = c.href
				}
				runs = append(runs, run)
			}
		}
		kind["runs"] = runs
	}

	node := map[string]any{
		"kind":     kind,
		"style":    mapStyle(t.style),
		"children": []any{},
	}
	if t.href != "" {
		node["href"] = t.href
	}
	return node
}

// ImageNode represents an image element.
type ImageNode struct {
	src    string
	width  *float64
	height *float64
	style  *Style
	href   string
	alt    string
}

// Image creates a new image node.
func Image(src string) *ImageNode {
	return &ImageNode{src: src}
}

func (i *ImageNode) Width(w float64) *ImageNode   { i.width = &w; return i }
func (i *ImageNode) Height(h float64) *ImageNode  { i.height = &h; return i }
func (i *ImageNode) Style(s Style) *ImageNode     { i.style = &s; return i }
func (i *ImageNode) Href(href string) *ImageNode   { i.href = href; return i }
func (i *ImageNode) Alt(alt string) *ImageNode     { i.alt = alt; return i }

func (i *ImageNode) toDict() map[string]any {
	kind := map[string]any{"type": "Image", "src": i.src}
	if i.width != nil {
		kind["width"] = *i.width
	}
	if i.height != nil {
		kind["height"] = *i.height
	}
	node := map[string]any{
		"kind":     kind,
		"style":    mapStyle(i.style),
		"children": []any{},
	}
	if i.href != "" {
		node["href"] = i.href
	}
	if i.alt != "" {
		node["alt"] = i.alt
	}
	return node
}

// TableNode represents a table.
type TableNode struct {
	children []Node
	columns  []Column
	style    *Style
}

// Table creates a new table node.
func Table(children ...Node) *TableNode {
	return &TableNode{children: children}
}

func (t *TableNode) Columns(cols []Column) *TableNode { t.columns = cols; return t }
func (t *TableNode) Style(s Style) *TableNode         { t.style = &s; return t }

func (t *TableNode) toDict() map[string]any {
	cols := make([]map[string]any, 0)
	for _, col := range t.columns {
		var mappedW any
		switch w := col.Width.(type) {
		case float64:
			mappedW = map[string]any{"Fixed": w}
		case int:
			mappedW = map[string]any{"Fixed": float64(w)}
		case string:
			if strings.HasSuffix(w, "fr") {
				f, _ := strconv.ParseFloat(w[:len(w)-2], 64)
				mappedW = map[string]any{"Fraction": f}
			} else if w == "auto" {
				mappedW = "Auto"
			} else {
				mappedW = "Auto"
			}
		default:
			mappedW = "Auto"
		}
		cols = append(cols, map[string]any{"width": mappedW})
	}
	return map[string]any{
		"kind":     map[string]any{"type": "Table", "columns": cols},
		"style":    mapStyle(t.style),
		"children": nodesToDicts(t.children),
	}
}

// RowNode represents a table row.
type RowNode struct {
	children []Node
	header   bool
	style    *Style
}

// Row creates a new table row.
func Row(children ...Node) *RowNode {
	return &RowNode{children: children}
}

func (r *RowNode) Header(h bool) *RowNode { r.header = h; return r }
func (r *RowNode) Style(s Style) *RowNode { r.style = &s; return r }

func (r *RowNode) toDict() map[string]any {
	return map[string]any{
		"kind":     map[string]any{"type": "TableRow", "is_header": r.header},
		"style":    mapStyle(r.style),
		"children": nodesToDicts(r.children),
	}
}

// CellNode represents a table cell.
type CellNode struct {
	children []Node
	colSpan  int
	rowSpan  int
	style    *Style
}

// Cell creates a new table cell.
func Cell(children ...Node) *CellNode {
	return &CellNode{children: children}
}

func (c *CellNode) ColSpan(n int) *CellNode { c.colSpan = n; return c }
func (c *CellNode) RowSpan(n int) *CellNode { c.rowSpan = n; return c }
func (c *CellNode) Style(s Style) *CellNode { c.style = &s; return c }

func (c *CellNode) toDict() map[string]any {
	kind := map[string]any{"type": "TableCell"}
	if c.colSpan != 0 {
		kind["col_span"] = c.colSpan
	}
	if c.rowSpan != 0 {
		kind["row_span"] = c.rowSpan
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(c.style),
		"children": nodesToDicts(c.children),
	}
}

// QRCodeNode represents a QR code element.
type QRCodeNode struct {
	data  string
	size  *float64
	color string
	style *Style
}

// QRCode creates a new QR code node.
func QRCode(data string) *QRCodeNode {
	return &QRCodeNode{data: data}
}

func (q *QRCodeNode) Size(size float64) *QRCodeNode  { q.size = &size; return q }
func (q *QRCodeNode) Color(c string) *QRCodeNode     { q.color = c; return q }
func (q *QRCodeNode) Style(s Style) *QRCodeNode      { q.style = &s; return q }

func (q *QRCodeNode) toDict() map[string]any {
	kind := map[string]any{"type": "QrCode", "data": q.data}
	if q.size != nil {
		kind["size"] = *q.size
	}
	if q.color != "" {
		kind["color"] = parseColor(q.color)
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(q.style),
		"children": []any{},
	}
}

// BarcodeNode represents a 1D barcode element.
type BarcodeNode struct {
	data   string
	format string
	width  *float64
	height float64
	color  string
	style  *Style
}

// Barcode creates a new barcode node.
func Barcode(data string) *BarcodeNode {
	return &BarcodeNode{data: data, format: "Code128", height: 60.0}
}

func (b *BarcodeNode) Format(f string) *BarcodeNode   { b.format = f; return b }
func (b *BarcodeNode) Width(w float64) *BarcodeNode    { b.width = &w; return b }
func (b *BarcodeNode) Height(h float64) *BarcodeNode   { b.height = h; return b }
func (b *BarcodeNode) Color(c string) *BarcodeNode     { b.color = c; return b }
func (b *BarcodeNode) Style(s Style) *BarcodeNode      { b.style = &s; return b }

func (b *BarcodeNode) toDict() map[string]any {
	kind := map[string]any{
		"type":   "Barcode",
		"data":   b.data,
		"format": b.format,
		"height": b.height,
	}
	if b.width != nil {
		kind["width"] = *b.width
	}
	style := mapStyle(b.style)
	if b.color != "" {
		style["color"] = parseColor(b.color)
	}
	return map[string]any{
		"kind":     kind,
		"style":    style,
		"children": []any{},
	}
}

// PageBreakNode forces a page break.
type PageBreakNode struct{}

// PageBreak creates a new page break node.
func PageBreak() *PageBreakNode { return &PageBreakNode{} }

func (p *PageBreakNode) toDict() map[string]any {
	return map[string]any{
		"kind":     map[string]any{"type": "PageBreak"},
		"style":    map[string]any{},
		"children": []any{},
	}
}

// FixedNode represents a fixed-position element (header/footer).
type FixedNode struct {
	children []Node
	position string
	style    *Style
}

// Fixed creates a new fixed node.
func Fixed(children ...Node) *FixedNode {
	return &FixedNode{children: children, position: "top"}
}

func (f *FixedNode) Position(pos string) *FixedNode { f.position = pos; return f }
func (f *FixedNode) Style(s Style) *FixedNode       { f.style = &s; return f }

func (f *FixedNode) toDict() map[string]any {
	posMap := map[string]string{"top": "Top", "bottom": "Bottom"}
	pos := "Top"
	if v, ok := posMap[f.position]; ok {
		pos = v
	}
	return map[string]any{
		"kind":     map[string]any{"type": "Fixed", "position": pos},
		"style":    mapStyle(f.style),
		"children": nodesToDicts(f.children),
	}
}

// WatermarkNode represents a watermark rendered on every page.
type WatermarkNode struct {
	text     string
	fontSize *float64
	color    string
	angle    *float64
	style    *Style
}

// Watermark creates a new watermark node.
func Watermark(text string) *WatermarkNode {
	return &WatermarkNode{text: text}
}

func (w *WatermarkNode) FontSize(fs float64) *WatermarkNode { w.fontSize = &fs; return w }
func (w *WatermarkNode) Color(c string) *WatermarkNode      { w.color = c; return w }
func (w *WatermarkNode) Angle(a float64) *WatermarkNode     { w.angle = &a; return w }
func (w *WatermarkNode) Style(s Style) *WatermarkNode       { w.style = &s; return w }

func (w *WatermarkNode) toDict() map[string]any {
	kind := map[string]any{"type": "Watermark", "text": w.text}
	if w.fontSize != nil {
		kind["font_size"] = *w.fontSize
	}
	if w.color != "" {
		kind["color"] = parseColor(w.color)
	}
	if w.angle != nil {
		kind["angle"] = *w.angle
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(w.style),
		"children": []any{},
	}
}

// ── Chart components ────────────────────────────────────────────────

// BarChartNode represents a bar chart.
type BarChartNode struct {
	data       []ChartDataPoint
	width      float64
	height     float64
	color      string
	showLabels bool
	showValues bool
	showGrid   bool
	title      string
	style      *Style
}

// BarChart creates a new bar chart node.
func BarChart(data []ChartDataPoint) *BarChartNode {
	return &BarChartNode{data: data, width: 400, height: 200, showLabels: true}
}

func (b *BarChartNode) Width(w float64) *BarChartNode       { b.width = w; return b }
func (b *BarChartNode) Height(h float64) *BarChartNode      { b.height = h; return b }
func (b *BarChartNode) Color(c string) *BarChartNode        { b.color = c; return b }
func (b *BarChartNode) ShowLabels(v bool) *BarChartNode     { b.showLabels = v; return b }
func (b *BarChartNode) ShowValues(v bool) *BarChartNode     { b.showValues = v; return b }
func (b *BarChartNode) ShowGrid(v bool) *BarChartNode       { b.showGrid = v; return b }
func (b *BarChartNode) Title(t string) *BarChartNode        { b.title = t; return b }
func (b *BarChartNode) Style(s Style) *BarChartNode         { b.style = &s; return b }

func (b *BarChartNode) toDict() map[string]any {
	kind := map[string]any{
		"type":        "BarChart",
		"data":        b.data,
		"width":       b.width,
		"height":      b.height,
		"show_labels": b.showLabels,
		"show_values": b.showValues,
		"show_grid":   b.showGrid,
	}
	if b.color != "" {
		kind["color"] = b.color
	}
	if b.title != "" {
		kind["title"] = b.title
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(b.style),
		"children": []any{},
	}
}

// LineChartNode represents a line chart.
type LineChartNode struct {
	series     []ChartSeries
	labels     []string
	width      float64
	height     float64
	showPoints bool
	showGrid   bool
	title      string
	style      *Style
}

// LineChart creates a new line chart node.
func LineChart(series []ChartSeries, labels []string) *LineChartNode {
	return &LineChartNode{series: series, labels: labels, width: 400, height: 200}
}

func (l *LineChartNode) Width(w float64) *LineChartNode       { l.width = w; return l }
func (l *LineChartNode) Height(h float64) *LineChartNode      { l.height = h; return l }
func (l *LineChartNode) ShowPoints(v bool) *LineChartNode     { l.showPoints = v; return l }
func (l *LineChartNode) ShowGrid(v bool) *LineChartNode       { l.showGrid = v; return l }
func (l *LineChartNode) Title(t string) *LineChartNode        { l.title = t; return l }
func (l *LineChartNode) Style(s Style) *LineChartNode         { l.style = &s; return l }

func (l *LineChartNode) toDict() map[string]any {
	kind := map[string]any{
		"type":        "LineChart",
		"series":      l.series,
		"labels":      l.labels,
		"width":       l.width,
		"height":      l.height,
		"show_points": l.showPoints,
		"show_grid":   l.showGrid,
	}
	if l.title != "" {
		kind["title"] = l.title
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(l.style),
		"children": []any{},
	}
}

// PieChartNode represents a pie/donut chart.
type PieChartNode struct {
	data       []ChartDataPoint
	width      float64
	height     float64
	donut      bool
	showLegend bool
	title      string
	style      *Style
}

// PieChart creates a new pie chart node.
func PieChart(data []ChartDataPoint) *PieChartNode {
	return &PieChartNode{data: data, width: 200, height: 200}
}

func (p *PieChartNode) Width(w float64) *PieChartNode       { p.width = w; return p }
func (p *PieChartNode) Height(h float64) *PieChartNode      { p.height = h; return p }
func (p *PieChartNode) Donut(v bool) *PieChartNode          { p.donut = v; return p }
func (p *PieChartNode) ShowLegend(v bool) *PieChartNode     { p.showLegend = v; return p }
func (p *PieChartNode) Title(t string) *PieChartNode        { p.title = t; return p }
func (p *PieChartNode) Style(s Style) *PieChartNode         { p.style = &s; return p }

func (p *PieChartNode) toDict() map[string]any {
	kind := map[string]any{
		"type":        "PieChart",
		"data":        p.data,
		"width":       p.width,
		"height":      p.height,
		"donut":       p.donut,
		"show_legend": p.showLegend,
	}
	if p.title != "" {
		kind["title"] = p.title
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(p.style),
		"children": []any{},
	}
}

// AreaChartNode represents an area chart.
type AreaChartNode struct {
	series   []ChartSeries
	labels   []string
	width    float64
	height   float64
	showGrid bool
	title    string
	style    *Style
}

// AreaChart creates a new area chart node.
func AreaChart(series []ChartSeries, labels []string) *AreaChartNode {
	return &AreaChartNode{series: series, labels: labels, width: 400, height: 200}
}

func (a *AreaChartNode) Width(w float64) *AreaChartNode   { a.width = w; return a }
func (a *AreaChartNode) Height(h float64) *AreaChartNode  { a.height = h; return a }
func (a *AreaChartNode) ShowGrid(v bool) *AreaChartNode   { a.showGrid = v; return a }
func (a *AreaChartNode) Title(t string) *AreaChartNode    { a.title = t; return a }
func (a *AreaChartNode) Style(s Style) *AreaChartNode     { a.style = &s; return a }

func (a *AreaChartNode) toDict() map[string]any {
	kind := map[string]any{
		"type":      "AreaChart",
		"series":    a.series,
		"labels":    a.labels,
		"width":     a.width,
		"height":    a.height,
		"show_grid": a.showGrid,
	}
	if a.title != "" {
		kind["title"] = a.title
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(a.style),
		"children": []any{},
	}
}

// DotPlotNode represents a dot plot (scatter plot).
type DotPlotNode struct {
	groups     []DotPlotGroup
	width      float64
	height     float64
	xMin       *float64
	xMax       *float64
	yMin       *float64
	yMax       *float64
	xLabel     string
	yLabel     string
	showLegend bool
	dotSize    float64
	style      *Style
}

// DotPlot creates a new dot plot node.
func DotPlot(groups []DotPlotGroup) *DotPlotNode {
	return &DotPlotNode{groups: groups, width: 400, height: 300, dotSize: 4.0}
}

func (d *DotPlotNode) Width(w float64) *DotPlotNode       { d.width = w; return d }
func (d *DotPlotNode) Height(h float64) *DotPlotNode      { d.height = h; return d }
func (d *DotPlotNode) XMin(v float64) *DotPlotNode        { d.xMin = &v; return d }
func (d *DotPlotNode) XMax(v float64) *DotPlotNode        { d.xMax = &v; return d }
func (d *DotPlotNode) YMin(v float64) *DotPlotNode        { d.yMin = &v; return d }
func (d *DotPlotNode) YMax(v float64) *DotPlotNode        { d.yMax = &v; return d }
func (d *DotPlotNode) XLabel(l string) *DotPlotNode       { d.xLabel = l; return d }
func (d *DotPlotNode) YLabel(l string) *DotPlotNode       { d.yLabel = l; return d }
func (d *DotPlotNode) ShowLegend(v bool) *DotPlotNode     { d.showLegend = v; return d }
func (d *DotPlotNode) DotSize(s float64) *DotPlotNode     { d.dotSize = s; return d }
func (d *DotPlotNode) Style(s Style) *DotPlotNode         { d.style = &s; return d }

func (d *DotPlotNode) toDict() map[string]any {
	kind := map[string]any{
		"type":        "DotPlot",
		"groups":      d.groups,
		"width":       d.width,
		"height":      d.height,
		"show_legend": d.showLegend,
		"dot_size":    d.dotSize,
	}
	if d.xMin != nil {
		kind["x_min"] = *d.xMin
	}
	if d.xMax != nil {
		kind["x_max"] = *d.xMax
	}
	if d.yMin != nil {
		kind["y_min"] = *d.yMin
	}
	if d.yMax != nil {
		kind["y_max"] = *d.yMax
	}
	if d.xLabel != "" {
		kind["x_label"] = d.xLabel
	}
	if d.yLabel != "" {
		kind["y_label"] = d.yLabel
	}
	return map[string]any{
		"kind":     kind,
		"style":    mapStyle(d.style),
		"children": []any{},
	}
}

// ── Helpers ─────────────────────────────────────────────────────────

func nodesToDicts(nodes []Node) []any {
	result := make([]any, len(nodes))
	for i, n := range nodes {
		result[i] = n.toDict()
	}
	return result
}

