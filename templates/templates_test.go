package templates

import (
	"encoding/json"
	"math"
	"testing"
)

// ── Document serialization ──────────────────────────────────────────

func TestDocumentToJSON(t *testing.T) {
	doc := Document(Page(Text("Hello")))
	j, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(j), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := parsed["children"]; !ok {
		t.Fatal("missing 'children' key")
	}
}

func TestDocumentWithMetadata(t *testing.T) {
	doc := Document(Page(Text("Hello"))).Title("Test").Author("Author")
	d := doc.toDict()
	meta := d["metadata"].(map[string]any)
	if meta["title"] != "Test" {
		t.Fatalf("expected title 'Test', got %v", meta["title"])
	}
	if meta["author"] != "Author" {
		t.Fatalf("expected author 'Author', got %v", meta["author"])
	}
}

func TestDocumentWithLang(t *testing.T) {
	doc := Document(Page(Text("Hello"))).Lang("en-US")
	d := doc.toDict()
	meta := d["metadata"].(map[string]any)
	if meta["lang"] != "en-US" {
		t.Fatalf("expected lang 'en-US', got %v", meta["lang"])
	}
}

func TestDocumentWithDefaultStyle(t *testing.T) {
	doc := Document(Page(Text("Hello"))).DefaultStyle(Style{FontSize: 14, FontFamily: "Inter"})
	d := doc.toDict()
	ds := d["default_style"].(map[string]any)
	if ds["fontSize"] != 14.0 {
		t.Fatalf("expected fontSize 14, got %v", ds["fontSize"])
	}
	if ds["fontFamily"] != "Inter" {
		t.Fatalf("expected fontFamily 'Inter', got %v", ds["fontFamily"])
	}
}

func TestDocumentTagged(t *testing.T) {
	doc := Document(Page(Text("Hello"))).Tagged(true)
	d := doc.toDict()
	if d["tagged"] != true {
		t.Fatal("expected tagged=true")
	}
}

func TestDocumentNoMetadataWhenEmpty(t *testing.T) {
	doc := Document(Page(Text("Hello")))
	d := doc.toDict()
	if _, ok := d["metadata"]; ok {
		t.Fatal("expected no metadata key when empty")
	}
}

// ── Page serialization ──────────────────────────────────────────────

func TestDefaultPage(t *testing.T) {
	page := Page(Text("Hello"))
	d := page.toDict()
	config := d["kind"].(map[string]any)["config"].(map[string]any)
	if config["size"] != "A4" {
		t.Fatalf("expected size 'A4', got %v", config["size"])
	}
	margin := config["margin"].(map[string]float64)
	if margin["top"] != 54 {
		t.Fatalf("expected margin top 54, got %v", margin["top"])
	}
}

func TestPageCustomSize(t *testing.T) {
	page := Page(Text("Hello")).Size("Letter")
	d := page.toDict()
	config := d["kind"].(map[string]any)["config"].(map[string]any)
	if config["size"] != "Letter" {
		t.Fatalf("expected size 'Letter', got %v", config["size"])
	}
}

func TestPageCustomSizeDict(t *testing.T) {
	page := Page(Text("Hello")).Size(map[string]any{"width": 400.0, "height": 600.0})
	d := page.toDict()
	config := d["kind"].(map[string]any)["config"].(map[string]any)
	size := config["size"].(map[string]any)
	custom := size["Custom"].(map[string]any)
	if custom["width"] != 400.0 {
		t.Fatalf("expected width 400, got %v", custom["width"])
	}
}

func TestPageCustomMargin(t *testing.T) {
	page := Page(Text("Hello")).Margin(36.0)
	d := page.toDict()
	config := d["kind"].(map[string]any)["config"].(map[string]any)
	margin := config["margin"].(map[string]float64)
	if margin["top"] != 36 || margin["right"] != 36 {
		t.Fatalf("expected uniform margin 36, got %v", margin)
	}
}

// ── View serialization ──────────────────────────────────────────────

func TestViewWithChildren(t *testing.T) {
	v := View(Text("Child 1"), Text("Child 2")).Style(Style{FlexDirection: "column"})
	d := v.toDict()
	if d["kind"].(map[string]any)["type"] != "View" {
		t.Fatal("expected type 'View'")
	}
	if d["style"].(map[string]any)["flexDirection"] != "Column" {
		t.Fatal("expected flexDirection 'Column'")
	}
	children := d["children"].([]any)
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
}

func TestViewWrap(t *testing.T) {
	v := View().Style(Style{Padding: 10.0}).Wrap(true)
	d := v.toDict()
	if d["style"].(map[string]any)["wrap"] != true {
		t.Fatal("expected wrap=true")
	}
}

func TestViewBookmark(t *testing.T) {
	v := View().Bookmark("section1")
	d := v.toDict()
	if d["bookmark"] != "section1" {
		t.Fatal("expected bookmark 'section1'")
	}
}

// ── Text serialization ──────────────────────────────────────────────

func TestSimpleText(t *testing.T) {
	txt := Text("Hello world")
	d := txt.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "Text" {
		t.Fatal("expected type 'Text'")
	}
	if kind["content"] != "Hello world" {
		t.Fatal("expected content 'Hello world'")
	}
	if _, ok := kind["runs"]; ok {
		t.Fatal("expected no runs on simple text")
	}
}

func TestTextWithStyle(t *testing.T) {
	txt := Text("Hello", Style{FontSize: 24})
	d := txt.toDict()
	style := d["style"].(map[string]any)
	if style["fontSize"] != 24.0 {
		t.Fatalf("expected fontSize 24, got %v", style["fontSize"])
	}
}

func TestTextWithRuns(t *testing.T) {
	txt := Text("").Children(
		"Normal text ",
		Text("bold text", Style{FontWeight: "bold"}),
	)
	d := txt.toDict()
	kind := d["kind"].(map[string]any)
	runs := kind["runs"].([]map[string]any)
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	if runs[0]["content"] != "Normal text " {
		t.Fatalf("expected first run content, got %v", runs[0]["content"])
	}
	if runs[1]["content"] != "bold text" {
		t.Fatalf("expected second run content, got %v", runs[1]["content"])
	}
	runStyle := runs[1]["style"].(map[string]any)
	if runStyle["fontWeight"] != 700 {
		t.Fatalf("expected fontWeight 700, got %v", runStyle["fontWeight"])
	}
}

func TestTextWithHref(t *testing.T) {
	txt := Text("Click me").Href("https://example.com")
	d := txt.toDict()
	if d["href"] != "https://example.com" {
		t.Fatal("expected href")
	}
}

// ── Image serialization ─────────────────────────────────────────────

func TestImage(t *testing.T) {
	img := Image("data:image/png;base64,abc").Width(100).Height(50)
	d := img.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "Image" {
		t.Fatal("expected type 'Image'")
	}
	if kind["src"] != "data:image/png;base64,abc" {
		t.Fatal("expected src")
	}
	if kind["width"] != 100.0 {
		t.Fatalf("expected width 100, got %v", kind["width"])
	}
	if kind["height"] != 50.0 {
		t.Fatalf("expected height 50, got %v", kind["height"])
	}
}

func TestImageWithAlt(t *testing.T) {
	img := Image("file.png").Alt("A photo")
	d := img.toDict()
	if d["alt"] != "A photo" {
		t.Fatal("expected alt text")
	}
}

// ── Table serialization ─────────────────────────────────────────────

func TestTable(t *testing.T) {
	tbl := Table(
		Row(Cell(Text("A")), Cell(Text("B"))).Header(true),
		Row(Cell(Text("1")), Cell(Text("2"))),
	).Columns([]Column{{Width: 100.0}, {Width: "1fr"}})
	d := tbl.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "Table" {
		t.Fatal("expected type 'Table'")
	}
	cols := kind["columns"].([]map[string]any)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	col0w := cols[0]["width"].(map[string]any)
	if col0w["Fixed"] != 100.0 {
		t.Fatalf("expected Fixed 100, got %v", col0w)
	}
	col1w := cols[1]["width"].(map[string]any)
	if col1w["Fraction"] != 1.0 {
		t.Fatalf("expected Fraction 1, got %v", col1w)
	}
	children := d["children"].([]any)
	if len(children) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(children))
	}
	row0 := children[0].(map[string]any)
	if row0["kind"].(map[string]any)["is_header"] != true {
		t.Fatal("expected first row to be header")
	}
}

// ── Misc components ─────────────────────────────────────────────────

func TestPageBreak(t *testing.T) {
	pb := PageBreak()
	d := pb.toDict()
	if d["kind"].(map[string]any)["type"] != "PageBreak" {
		t.Fatal("expected type 'PageBreak'")
	}
}

func TestQRCode(t *testing.T) {
	qr := QRCode("https://example.com").Size(100)
	d := qr.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "QrCode" {
		t.Fatal("expected type 'QrCode'")
	}
	if kind["data"] != "https://example.com" {
		t.Fatal("expected data")
	}
	if kind["size"] != 100.0 {
		t.Fatalf("expected size 100, got %v", kind["size"])
	}
}

func TestQRCodeWithColor(t *testing.T) {
	qr := QRCode("test").Size(80).Color("#ff0000")
	d := qr.toDict()
	kind := d["kind"].(map[string]any)
	color := kind["color"].(map[string]float64)
	if color["r"] != 1.0 {
		t.Fatalf("expected r=1.0, got %v", color["r"])
	}
	if color["g"] != 0.0 {
		t.Fatalf("expected g=0.0, got %v", color["g"])
	}
}

func TestQRCodeMinimal(t *testing.T) {
	qr := QRCode("data")
	d := qr.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "QrCode" {
		t.Fatal("expected type 'QrCode'")
	}
	if _, ok := kind["size"]; ok {
		t.Fatal("expected no size on minimal QR code")
	}
	if _, ok := kind["color"]; ok {
		t.Fatal("expected no color on minimal QR code")
	}
}

func TestQRCodeWithStyle(t *testing.T) {
	qr := QRCode("url").Size(120).Style(Style{Margin: 8.0})
	d := qr.toDict()
	style := d["style"].(map[string]any)
	margin := style["margin"].(map[string]float64)
	if margin["top"] != 8 {
		t.Fatalf("expected margin top 8, got %v", margin["top"])
	}
}

func TestFixed(t *testing.T) {
	f := Fixed(Text("Header")).Position("top")
	d := f.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "Fixed" {
		t.Fatal("expected type 'Fixed'")
	}
	if kind["position"] != "Top" {
		t.Fatal("expected position 'Top'")
	}
}

func TestWatermark(t *testing.T) {
	w := Watermark("DRAFT").FontSize(60).Color("rgba(0,0,0,0.1)").Angle(-45)
	d := w.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "Watermark" {
		t.Fatal("expected type 'Watermark'")
	}
	if kind["text"] != "DRAFT" {
		t.Fatal("expected text 'DRAFT'")
	}
	if kind["font_size"] != 60.0 {
		t.Fatalf("expected font_size 60, got %v", kind["font_size"])
	}
	if kind["angle"] != -45.0 {
		t.Fatalf("expected angle -45, got %v", kind["angle"])
	}
}

// ── Barcode ─────────────────────────────────────────────────────────

func TestBarcodeDefaultFormat(t *testing.T) {
	bc := Barcode("ABC-123")
	d := bc.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "Barcode" {
		t.Fatal("expected type 'Barcode'")
	}
	if kind["data"] != "ABC-123" {
		t.Fatal("expected data")
	}
	if kind["format"] != "Code128" {
		t.Fatalf("expected format Code128, got %v", kind["format"])
	}
	if kind["height"] != 60.0 {
		t.Fatalf("expected height 60, got %v", kind["height"])
	}
}

func TestBarcodeCustomFormat(t *testing.T) {
	bc := Barcode("HELLO").Format("Code39").Width(200).Height(40)
	d := bc.toDict()
	kind := d["kind"].(map[string]any)
	if kind["format"] != "Code39" {
		t.Fatalf("expected format Code39, got %v", kind["format"])
	}
	if kind["width"] != 200.0 {
		t.Fatalf("expected width 200, got %v", kind["width"])
	}
	if kind["height"] != 40.0 {
		t.Fatalf("expected height 40, got %v", kind["height"])
	}
}

func TestBarcodeWithColor(t *testing.T) {
	bc := Barcode("12345").Color("#003366")
	d := bc.toDict()
	style := d["style"].(map[string]any)
	if _, ok := style["color"]; !ok {
		t.Fatal("expected color in style")
	}
}

func TestBarcodeNoWidthOmitted(t *testing.T) {
	bc := Barcode("test")
	d := bc.toDict()
	kind := d["kind"].(map[string]any)
	if _, ok := kind["width"]; ok {
		t.Fatal("expected no width on default barcode")
	}
}

// ── Style mapping ───────────────────────────────────────────────────

func TestStyleFontWeightBold(t *testing.T) {
	result := mapStyle(&Style{FontWeight: "bold"})
	if result["fontWeight"] != 700 {
		t.Fatalf("expected 700, got %v", result["fontWeight"])
	}
}

func TestStyleFontWeightNormal(t *testing.T) {
	result := mapStyle(&Style{FontWeight: "normal"})
	if result["fontWeight"] != 400 {
		t.Fatalf("expected 400, got %v", result["fontWeight"])
	}
}

func TestStyleFontWeightNumber(t *testing.T) {
	result := mapStyle(&Style{FontWeight: 600})
	if result["fontWeight"] != 600 {
		t.Fatalf("expected 600, got %v", result["fontWeight"])
	}
}

func TestStyleFlexDirection(t *testing.T) {
	result := mapStyle(&Style{FlexDirection: "column"})
	if result["flexDirection"] != "Column" {
		t.Fatalf("expected 'Column', got %v", result["flexDirection"])
	}
}

func TestStyleFlexDirectionRowReverse(t *testing.T) {
	result := mapStyle(&Style{FlexDirection: "row-reverse"})
	if result["flexDirection"] != "RowReverse" {
		t.Fatalf("expected 'RowReverse', got %v", result["flexDirection"])
	}
}

func TestStyleJustifyContent(t *testing.T) {
	result := mapStyle(&Style{JustifyContent: "space-between"})
	if result["justifyContent"] != "SpaceBetween" {
		t.Fatalf("expected 'SpaceBetween', got %v", result["justifyContent"])
	}
}

func TestStyleAlignItems(t *testing.T) {
	result := mapStyle(&Style{AlignItems: "center"})
	if result["alignItems"] != "Center" {
		t.Fatalf("expected 'Center', got %v", result["alignItems"])
	}
}

func TestStyleFlexShorthand(t *testing.T) {
	result := mapStyle(&Style{Flex: 1})
	if result["flexGrow"] != 1.0 {
		t.Fatalf("expected flexGrow 1, got %v", result["flexGrow"])
	}
	if result["flexShrink"] != 1.0 {
		t.Fatalf("expected flexShrink 1, got %v", result["flexShrink"])
	}
	basis := result["flexBasis"].(map[string]any)
	if basis["Pt"] != 0.0 {
		t.Fatalf("expected flexBasis Pt:0, got %v", basis)
	}
}

func TestStyleTextAlign(t *testing.T) {
	result := mapStyle(&Style{TextAlign: "center"})
	if result["textAlign"] != "Center" {
		t.Fatalf("expected 'Center', got %v", result["textAlign"])
	}
}

func TestStyleTextDecoration(t *testing.T) {
	result := mapStyle(&Style{TextDecoration: "line-through"})
	if result["textDecoration"] != "LineThrough" {
		t.Fatalf("expected 'LineThrough', got %v", result["textDecoration"])
	}
}

func TestStyleDisplayGrid(t *testing.T) {
	result := mapStyle(&Style{Display: "grid"})
	if result["display"] != "Grid" {
		t.Fatalf("expected 'Grid', got %v", result["display"])
	}
}

func TestStylePositionAbsolute(t *testing.T) {
	result := mapStyle(&Style{Position: "absolute"})
	if result["position"] != "Absolute" {
		t.Fatalf("expected 'Absolute', got %v", result["position"])
	}
}

func TestStyleOverflowHidden(t *testing.T) {
	result := mapStyle(&Style{Overflow: "hidden"})
	if result["overflow"] != "Hidden" {
		t.Fatalf("expected 'Hidden', got %v", result["overflow"])
	}
}

func TestStyleEmptyIsEmpty(t *testing.T) {
	result := mapStyle(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
	result2 := mapStyle(&Style{})
	if len(result2) != 0 {
		t.Fatalf("expected empty map, got %v", result2)
	}
}

func TestStylePaddingEdges(t *testing.T) {
	result := mapStyle(&Style{Padding: 40.0})
	pad := result["padding"].(map[string]float64)
	if pad["top"] != 40 || pad["right"] != 40 || pad["bottom"] != 40 || pad["left"] != 40 {
		t.Fatalf("expected uniform padding 40, got %v", pad)
	}
}

func TestStylePaddingOverride(t *testing.T) {
	result := mapStyle(&Style{Padding: 10.0, PaddingTop: 20})
	pad := result["padding"].(map[string]float64)
	if pad["top"] != 20 {
		t.Fatalf("expected padding top 20, got %v", pad["top"])
	}
	if pad["right"] != 10 {
		t.Fatalf("expected padding right 10, got %v", pad["right"])
	}
}

func TestStyleBorderShorthand(t *testing.T) {
	result := mapStyle(&Style{Border: "1px solid #000"})
	bw := result["borderWidth"].(map[string]float64)
	if bw["top"] != 1 {
		t.Fatalf("expected borderWidth top 1, got %v", bw["top"])
	}
	bc := result["borderColor"].(map[string]any)
	topColor := bc["top"].(map[string]float64)
	if topColor["r"] != 0.0 {
		t.Fatalf("expected borderColor top r=0, got %v", topColor["r"])
	}
}

// ── Dimension mapping ───────────────────────────────────────────────

func TestDimensionNumber(t *testing.T) {
	result := mapDimension(200.0)
	m := result.(map[string]any)
	if m["Pt"] != 200.0 {
		t.Fatalf("expected Pt:200, got %v", m)
	}
}

func TestDimensionPercent(t *testing.T) {
	result := mapDimension("50%")
	m := result.(map[string]any)
	if m["Percent"] != 50.0 {
		t.Fatalf("expected Percent:50, got %v", m)
	}
}

func TestDimensionAuto(t *testing.T) {
	result := mapDimension("auto")
	if result != "Auto" {
		t.Fatalf("expected 'Auto', got %v", result)
	}
}

func TestDimensionInt(t *testing.T) {
	result := mapDimension(100)
	m := result.(map[string]any)
	if m["Pt"] != 100.0 {
		t.Fatalf("expected Pt:100, got %v", m)
	}
}

// ── Color parsing ───────────────────────────────────────────────────

func TestColorHex(t *testing.T) {
	c := parseColor("#ff0000")
	if math.Abs(c["r"]-1.0) > 0.01 {
		t.Fatalf("expected r~1.0, got %v", c["r"])
	}
	if math.Abs(c["g"]) > 0.01 {
		t.Fatalf("expected g~0.0, got %v", c["g"])
	}
	if c["a"] != 1.0 {
		t.Fatalf("expected a=1.0, got %v", c["a"])
	}
}

func TestColorHexShort(t *testing.T) {
	c := parseColor("#f00")
	if math.Abs(c["r"]-1.0) > 0.01 {
		t.Fatalf("expected r~1.0, got %v", c["r"])
	}
	if math.Abs(c["g"]) > 0.01 {
		t.Fatalf("expected g~0.0, got %v", c["g"])
	}
}

func TestColorRGBA(t *testing.T) {
	c := parseColor("rgba(255, 128, 0, 0.5)")
	if math.Abs(c["r"]-1.0) > 0.01 {
		t.Fatalf("expected r~1.0, got %v", c["r"])
	}
	if math.Abs(c["g"]-128.0/255.0) > 0.01 {
		t.Fatalf("expected g~0.502, got %v", c["g"])
	}
	if math.Abs(c["b"]) > 0.01 {
		t.Fatalf("expected b~0.0, got %v", c["b"])
	}
	if math.Abs(c["a"]-0.5) > 0.01 {
		t.Fatalf("expected a~0.5, got %v", c["a"])
	}
}

func TestColorRGB(t *testing.T) {
	c := parseColor("rgb(0, 255, 0)")
	if math.Abs(c["g"]-1.0) > 0.01 {
		t.Fatalf("expected g~1.0, got %v", c["g"])
	}
}

// ── Edge expansion ──────────────────────────────────────────────────

func TestExpandEdgesNumber(t *testing.T) {
	e := expandEdges(40.0)
	if e["top"] != 40 || e["right"] != 40 || e["bottom"] != 40 || e["left"] != 40 {
		t.Fatalf("expected uniform 40, got %v", e)
	}
}

func TestExpandEdgesSliceTwo(t *testing.T) {
	e := expandEdges([]float64{10, 20})
	if e["top"] != 10 || e["right"] != 20 || e["bottom"] != 10 || e["left"] != 20 {
		t.Fatalf("expected 10/20 pattern, got %v", e)
	}
}

func TestExpandEdgesSliceFour(t *testing.T) {
	e := expandEdges([]float64{1, 2, 3, 4})
	if e["top"] != 1 || e["right"] != 2 || e["bottom"] != 3 || e["left"] != 4 {
		t.Fatalf("expected 1/2/3/4, got %v", e)
	}
}

func TestExpandEdgesSliceThree(t *testing.T) {
	e := expandEdges([]float64{10, 20, 30})
	if e["top"] != 10 || e["right"] != 20 || e["bottom"] != 30 || e["left"] != 20 {
		t.Fatalf("expected 10/20/30/20, got %v", e)
	}
}

// ── Grid template parsing ───────────────────────────────────────────

func TestGridTemplateSimple(t *testing.T) {
	result := parseGridTemplate("1fr 200 auto")
	if len(result) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(result))
	}
	fr := result[0].(map[string]any)
	if fr["Fr"] != 1.0 {
		t.Fatalf("expected Fr:1, got %v", fr)
	}
	pt := result[1].(map[string]any)
	if pt["Pt"] != 200.0 {
		t.Fatalf("expected Pt:200, got %v", pt)
	}
	if result[2] != "Auto" {
		t.Fatalf("expected Auto, got %v", result[2])
	}
}

func TestGridTemplateRepeat(t *testing.T) {
	result := parseGridTemplate("repeat(3, 1fr)")
	if len(result) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(result))
	}
	for i, r := range result {
		fr := r.(map[string]any)
		if fr["Fr"] != 1.0 {
			t.Fatalf("track %d: expected Fr:1, got %v", i, fr)
		}
	}
}

func TestGridTemplateMixedRepeat(t *testing.T) {
	result := parseGridTemplate("200 repeat(2, 1fr) 200")
	if len(result) != 4 {
		t.Fatalf("expected 4 tracks, got %d", len(result))
	}
}

// ── Chart components ────────────────────────────────────────────────

func TestBarChart(t *testing.T) {
	bc := BarChart([]ChartDataPoint{
		{Label: "A", Value: 10},
		{Label: "B", Value: 20},
	}).Title("Sales")
	d := bc.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "BarChart" {
		t.Fatal("expected type 'BarChart'")
	}
	if kind["title"] != "Sales" {
		t.Fatalf("expected title 'Sales', got %v", kind["title"])
	}
	if kind["width"] != 400.0 {
		t.Fatalf("expected default width 400, got %v", kind["width"])
	}
}

func TestLineChart(t *testing.T) {
	lc := LineChart(
		[]ChartSeries{{Data: []float64{1, 2, 3}, Color: "#f00"}},
		[]string{"Jan", "Feb", "Mar"},
	).ShowPoints(true)
	d := lc.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "LineChart" {
		t.Fatal("expected type 'LineChart'")
	}
	if kind["show_points"] != true {
		t.Fatal("expected show_points=true")
	}
}

func TestPieChart(t *testing.T) {
	pc := PieChart([]ChartDataPoint{
		{Label: "A", Value: 30},
		{Label: "B", Value: 70},
	}).Donut(true).ShowLegend(true)
	d := pc.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "PieChart" {
		t.Fatal("expected type 'PieChart'")
	}
	if kind["donut"] != true {
		t.Fatal("expected donut=true")
	}
	if kind["show_legend"] != true {
		t.Fatal("expected show_legend=true")
	}
}

func TestAreaChart(t *testing.T) {
	ac := AreaChart(
		[]ChartSeries{{Data: []float64{1, 2, 3}}},
		[]string{"Q1", "Q2", "Q3"},
	).ShowGrid(true).Title("Revenue")
	d := ac.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "AreaChart" {
		t.Fatal("expected type 'AreaChart'")
	}
	if kind["show_grid"] != true {
		t.Fatal("expected show_grid=true")
	}
	if kind["title"] != "Revenue" {
		t.Fatal("expected title 'Revenue'")
	}
}

func TestDotPlot(t *testing.T) {
	dp := DotPlot([]DotPlotGroup{
		{Label: "Group A", Data: []DotPlotPoint{{X: 1, Y: 2}, {X: 3, Y: 4}}},
	}).XLabel("X").YLabel("Y").ShowLegend(true)
	d := dp.toDict()
	kind := d["kind"].(map[string]any)
	if kind["type"] != "DotPlot" {
		t.Fatal("expected type 'DotPlot'")
	}
	if kind["x_label"] != "X" {
		t.Fatal("expected x_label 'X'")
	}
	if kind["show_legend"] != true {
		t.Fatal("expected show_legend=true")
	}
}

// ── WASM integration tests ──────────────────────────────────────────

func skipIfNoWasm(t *testing.T) {
	t.Helper()
	_, err := renderPDF(`{"children":[]}`)
	if err != nil {
		t.Skipf("WASM not available: %v", err)
	}
}

func TestWASMSimpleDocument(t *testing.T) {
	skipIfNoWasm(t)

	doc := Document(Page(Text("Hello World")))
	pdf, err := doc.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(pdf) < 4 || string(pdf[:4]) != "%PDF" {
		t.Fatal("result is not a valid PDF")
	}
}

func TestWASMEmbedData(t *testing.T) {
	skipIfNoWasm(t)

	doc := Document(Page(Text("Hello")))
	pdf, err := doc.Render(map[string]any{"customer": "Acme"})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(pdf) < 4 || string(pdf[:4]) != "%PDF" {
		t.Fatal("result is not a valid PDF")
	}
}

func TestWASMTable(t *testing.T) {
	skipIfNoWasm(t)

	doc := Document(Page(
		Table(
			Row(Cell(Text("Name")), Cell(Text("Value"))).Header(true),
			Row(Cell(Text("A")), Cell(Text("1"))),
		).Columns([]Column{{Width: "1fr"}, {Width: "1fr"}}),
	))
	pdf, err := doc.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(pdf) < 4 || string(pdf[:4]) != "%PDF" {
		t.Fatal("result is not a valid PDF")
	}
}

func TestWASMBarChart(t *testing.T) {
	skipIfNoWasm(t)

	doc := Document(Page(
		BarChart([]ChartDataPoint{
			{Label: "A", Value: 10},
			{Label: "B", Value: 20},
		}),
	))
	pdf, err := doc.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(pdf) < 4 || string(pdf[:4]) != "%PDF" {
		t.Fatal("result is not a valid PDF")
	}
}
