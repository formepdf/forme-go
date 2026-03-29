# Forme Go SDK

Go SDK for [Forme](https://formepdf.com) — a PDF rendering engine. Two packages:

- **`forme-go`** — Zero-dependency API client for the hosted rendering service
- **`forme-go/templates`** — Local PDF rendering with a component DSL and WASM engine

## Installation

```bash
# API client only (zero dependencies)
go get github.com/formepdf/forme-go

# Native templates (adds wazero dependency)
go get github.com/formepdf/forme-go/templates
```

## Quick Start: Hosted API

```go
package main

import (
    "os"
    forme "github.com/formepdf/forme-go"
)

func main() {
    client := forme.New(os.Getenv("FORME_API_KEY"))

    // Render a template to PDF
    pdf, err := client.Render("invoice", map[string]any{
        "customer": "Acme Corp",
        "total":    150.00,
    })
    if err != nil {
        panic(err)
    }
    os.WriteFile("invoice.pdf", pdf, 0644)

    // Extract embedded data from a PDF
    data, err := client.Extract(pdf)
    if err != nil {
        panic(err)
    }
    // data == map[string]any{"customer": "Acme Corp", "total": 150.00}
}
```

## Quick Start: Native Templates

```go
package main

import (
    "os"
    t "github.com/formepdf/forme-go/templates"
)

func main() {
    doc := t.Document(
        t.Page(
            t.View(
                t.Text("Invoice", t.Style{FontSize: 24, FontWeight: "bold"}),
                t.Text("Acme Corp", t.Style{FontSize: 14, Color: "#666"}),
            ).Style(t.Style{FlexDirection: "column", Gap: 8}),

            t.Table(
                t.Row(t.Cell(t.Text("Item")), t.Cell(t.Text("Price"))).Header(true),
                t.Row(t.Cell(t.Text("Widget")), t.Cell(t.Text("$50.00"))),
                t.Row(t.Cell(t.Text("Gadget")), t.Cell(t.Text("$100.00"))),
            ).Columns([]t.Column{{Width: "1fr"}, {Width: 100.0}}),
        ),
    ).Title("Invoice #001")

    // Serialize to JSON (no WASM needed)
    json, _ := doc.ToJSON()

    // Render to PDF (requires WASM — build with -tags forme_wasm)
    pdf, err := doc.Render()
    if err != nil {
        panic(err)
    }
    os.WriteFile("invoice.pdf", pdf, 0644)

    _ = json
}
```

## API Client Reference

### Constructor

```go
client := forme.New(apiKey string, opts ...forme.Option)
```

| Option | Description |
|--------|-------------|
| `forme.WithBaseURL(url)` | Custom API base URL (default: `https://api.formepdf.com`) |
| `forme.WithHTTPClient(c)` | Custom `*http.Client` |

### Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `Render(slug, data)` | Synchronous render | `([]byte, error)` — PDF bytes |
| `RenderWithOptions(slug, data, opts)` | Render with options | `([]byte, error)` |
| `RenderS3(slug, data, s3Opts)` | Render and upload to S3 | `(*S3Result, error)` |
| `RenderAsync(slug, data, asyncOpts)` | Start async render job | `(*AsyncResult, error)` |
| `GetJob(jobID)` | Poll async job status | `(*JobResult, error)` |
| `Merge(pdfs)` | Merge multiple PDFs | `([]byte, error)` |
| `Extract(pdfBytes)` | Extract embedded data | `(map[string]any, error)` |

### Error Handling

All methods return `*forme.FormeError` on non-2xx responses:

```go
pdf, err := client.Render("invoice", data)
if err != nil {
    var fErr *forme.FormeError
    if errors.As(err, &fErr) {
        fmt.Printf("API error %d: %s\n", fErr.Status, fErr.Message)
    }
}
```

`Extract` returns `nil, nil` when the PDF has no embedded data (404 with "no embedded data").

## Component Reference

| Constructor | Description | Chainable Methods |
|-------------|-------------|-------------------|
| `Document(children...)` | Root document | `.Title()`, `.Author()`, `.Subject()`, `.Lang()`, `.DefaultStyle()`, `.Fonts()`, `.Tagged()` |
| `Page(children...)` | Page container | `.Size()`, `.Margin()` |
| `View(children...)` | Flex/grid container | `.Style()`, `.Wrap()`, `.Bookmark()`, `.Href()` |
| `Text(content, style?)` | Text element | `.Style()`, `.Href()`, `.Children()` |
| `Image(src)` | Image element | `.Width()`, `.Height()`, `.Style()`, `.Href()`, `.Alt()` |
| `Table(children...)` | Table container | `.Columns()`, `.Style()` |
| `Row(children...)` | Table row | `.Header()`, `.Style()` |
| `Cell(children...)` | Table cell | `.ColSpan()`, `.RowSpan()`, `.Style()` |
| `QRCode(data)` | QR code | `.Size()`, `.Color()`, `.Style()` |
| `Barcode(data)` | 1D barcode | `.Format()`, `.Width()`, `.Height()`, `.Color()`, `.Style()` |
| `PageBreak()` | Force page break | — |
| `Fixed(children...)` | Fixed header/footer | `.Position()`, `.Style()` |
| `Watermark(text)` | Page watermark | `.FontSize()`, `.Color()`, `.Angle()`, `.Style()` |
| `BarChart(data)` | Bar chart | `.Width()`, `.Height()`, `.Color()`, `.ShowLabels()`, `.ShowValues()`, `.ShowGrid()`, `.Title()`, `.Style()` |
| `LineChart(series, labels)` | Line chart | `.Width()`, `.Height()`, `.ShowPoints()`, `.ShowGrid()`, `.Title()`, `.Style()` |
| `PieChart(data)` | Pie/donut chart | `.Width()`, `.Height()`, `.Donut()`, `.ShowLegend()`, `.Title()`, `.Style()` |
| `AreaChart(series, labels)` | Area chart | `.Width()`, `.Height()`, `.ShowGrid()`, `.Title()`, `.Style()` |
| `DotPlot(groups)` | Scatter plot | `.Width()`, `.Height()`, `.XMin()`, `.XMax()`, `.YMin()`, `.YMax()`, `.XLabel()`, `.YLabel()`, `.ShowLegend()`, `.DotSize()`, `.Style()` |

## Style Properties

The `Style` struct supports all CSS-like properties:

**Dimensions**: `Width`, `Height`, `MinWidth`, `MinHeight`, `MaxWidth`, `MaxHeight` — accepts `float64`, `"auto"`, `"50%"`

**Spacing**: `Padding`, `PaddingTop/Right/Bottom/Left`, `PaddingHorizontal/Vertical`, `Margin` (same pattern)

**Flex**: `Flex`, `FlexDirection` (`"row"`, `"column"`, `"row-reverse"`, `"column-reverse"`), `JustifyContent`, `AlignItems`, `AlignSelf`, `AlignContent`, `FlexWrap`, `FlexGrow`, `FlexShrink`, `FlexBasis`, `Gap`, `RowGap`, `ColumnGap`

**Grid**: `Display` (`"grid"`), `GridTemplateColumns` (e.g. `"1fr 200 auto"`, `"repeat(3, 1fr)"`), `GridTemplateRows`, `GridAutoRows`, `GridAutoColumns`, `GridColumnStart/End`, `GridRowStart/End`, `GridColumnSpan`, `GridRowSpan`

**Typography**: `FontFamily`, `FontSize`, `FontWeight` (`"bold"`, `"normal"`, or int), `FontStyle`, `LineHeight`, `TextAlign`, `LetterSpacing`, `TextDecoration`, `TextTransform`, `Hyphens`, `Lang`, `Direction`, `TextOverflow`, `LineBreaking`

**Color**: `Color`, `BackgroundColor` (hex, rgb(), rgba()), `Opacity`

**Border**: `Border` (shorthand: `"1px solid #000"`), `BorderTop/Right/Bottom/Left`, `BorderWidth`, `BorderColor`, per-side variants, `BorderRadius` and per-corner variants

**Position**: `Position` (`"relative"`, `"absolute"`), `Top`, `Right`, `Bottom`, `Left`

**Overflow**: `Overflow` (`"visible"`, `"hidden"`)

## Hosted vs Native

| | Hosted API | Native Templates |
|---|---|---|
| **Dependency** | Zero (stdlib only) | wazero (~10MB WASM) |
| **Templates** | Pre-uploaded via dashboard | Built in Go code |
| **Network** | Requires API call | Fully offline |
| **Use case** | Dynamic data + stored templates | Full control, CI/CD, testing |
| **Rendering** | Server-side | Local WASM |

## Building WASM for Local Rendering

```bash
cd packages/go-sdk/templates
bash build_wasm.sh
go test -tags forme_wasm ./...
```

The WASM binary is `.gitignore`d — build it locally or download from releases.
