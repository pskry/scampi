// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Parsing
// -----------------------------------------------------------------------------

var (
	benchRe = regexp.MustCompile(
		`^Benchmark([^/\s]+)(?:/Size-(\d+))?-\d+\s+\d+\s+([\d.]+)\s+ns/op`,
	)
	tsRe = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})T(\d{2})(\d{2})`)
)

type groupKey struct {
	family string
	size   string
	ts     string
}

type point struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type series struct {
	Size   string  `json:"size"`
	Points []point `json:"points"`
}

type family struct {
	Name   string   `json:"name"`
	Series []series `json:"series"`
}

type chartData struct {
	Dates    []string `json:"dates"`
	Families []family `json:"families"`
}

func main() {
	outPath := flag.String("o", "bench.html", "output HTML file path")
	flag.Parse()

	if flag.NArg() < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: benchplot [-o out.html] <benchmark files...>")
		os.Exit(1)
	}

	values := make(map[groupKey][]float64)
	for _, path := range flag.Args() {
		ts := extractTimestamp(path)
		if ts == "" {
			continue
		}
		parseFile(path, ts, values)
	}

	data := buildChartData(values)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "json marshal: %v\n", err)
		os.Exit(1)
	}

	if err := writeHTML(*outPath, string(jsonBytes)); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "write html: %v\n", err)
		os.Exit(1)
	}
}

func extractTimestamp(path string) string {
	base := filepath.Base(path)
	m := tsRe.FindStringSubmatch(base)
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%sT%s:%s", m[1], m[2], m[3])
}

func parseFile(path, ts string, out map[groupKey][]float64) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		m := benchRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		familyName := m[1]
		size := m[2]
		if size == "" {
			size = "1"
		}
		ns, _ := strconv.ParseFloat(m[3], 64)

		k := groupKey{family: familyName, size: size, ts: ts}
		out[k] = append(out[k], ns)
	}
}

func buildChartData(values map[groupKey][]float64) chartData {
	familyMap := make(map[string]map[string]map[string][]float64)
	dateSet := make(map[string]bool)
	for k, vs := range values {
		dateSet[k.ts] = true
		if familyMap[k.family] == nil {
			familyMap[k.family] = make(map[string]map[string][]float64)
		}
		if familyMap[k.family][k.size] == nil {
			familyMap[k.family][k.size] = make(map[string][]float64)
		}
		familyMap[k.family][k.size][k.ts] = vs
	}

	var dates []string
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	var families []family
	for name, sizes := range familyMap {
		var seriesList []series
		for size, timestamps := range sizes {
			var pts []point
			for ts, vs := range timestamps {
				med := median(vs)
				pts = append(pts, point{Date: ts, Value: med / 1e9})
			}
			sort.Slice(pts, func(i, j int) bool {
				return pts[i].Date < pts[j].Date
			})
			seriesList = append(seriesList, series{Size: size, Points: pts})
		}
		sort.Slice(seriesList, func(i, j int) bool {
			a, _ := strconv.Atoi(seriesList[i].Size)
			b, _ := strconv.Atoi(seriesList[j].Size)
			return a < b
		})
		families = append(families, family{Name: name, Series: seriesList})
	}
	sort.Slice(families, func(i, j int) bool {
		return families[i].Name < families[j].Name
	})

	return chartData{Dates: dates, Families: families}
}

func median(xs []float64) float64 {
	slices.Sort(xs)
	n := len(xs)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return xs[n/2]
	}
	return (xs[n/2-1] + xs[n/2]) / 2
}

// HTML output
// -----------------------------------------------------------------------------

func writeHTML(path, jsonData string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return htmlTmpl.Execute(f, template.JS(jsonData)) //nolint:gosec // trusted data from local benchmark files
}

var htmlTmpl = template.Must(template.New("dashboard").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Benchmark Dashboard</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, monospace;
    background: #0d1117; color: #c9d1d9; padding: 24px;
  }
  h1 { font-size: 20px; margin-bottom: 24px; color: #58a6ff; }
  .chart-container {
    background: #161b22; border: 1px solid #30363d; border-radius: 6px;
    padding: 20px; margin-bottom: 24px;
  }
  .chart-title { font-size: 15px; font-weight: 600; margin-bottom: 12px; color: #e6edf3; }
  .chart-area { position: relative; }
  svg { display: block; }
  .legend {
    display: flex; flex-wrap: wrap; gap: 6px 16px;
    margin-top: 10px; font-size: 12px;
  }
  .legend-item {
    cursor: pointer; display: flex; align-items: center; gap: 5px;
    user-select: none; padding: 2px 6px; border-radius: 3px;
  }
  .legend-item:hover { background: #21262d; }
  .legend-item.hidden { opacity: 0.35; }
  .legend-swatch {
    width: 12px; height: 3px; border-radius: 1px; flex-shrink: 0;
  }
  .tooltip {
    position: fixed; pointer-events: none; background: #1c2128;
    border: 1px solid #444c56; border-radius: 4px; padding: 6px 10px;
    font-size: 12px; line-height: 1.5; white-space: nowrap;
    display: none; z-index: 100; color: #e6edf3;
  }
  .tooltip-label { color: #8b949e; }
</style>
</head>
<body>
<h1>Benchmark Dashboard</h1>
<div id="charts"></div>
<div class="tooltip" id="tooltip"></div>
<script>
"use strict";
const DATA = {{.}};
const COLORS = ["#58a6ff","#f0883e","#3fb950","#bc8cff","#f778ba","#79c0ff","#d2a8ff","#ffd33d"];
const MARGIN = {top: 10, right: 20, bottom: 50, left: 72};
const WIDTH = 780;
const HEIGHT = 280;
const innerW = WIDTH - MARGIN.left - MARGIN.right;
const innerH = HEIGHT - MARGIN.top - MARGIN.bottom;

const tooltip = document.getElementById("tooltip");
const chartsDiv = document.getElementById("charts");

function formatValue(v) {
  if (v >= 1) return v.toFixed(3) + " s";
  if (v >= 1e-3) return (v * 1e3).toFixed(3) + " ms";
  if (v >= 1e-6) return (v * 1e6).toFixed(3) + " µs";
  return (v * 1e9).toFixed(1) + " ns";
}

function formatDate(d) {
  return d.replace("T", " ");
}

function logTicks(lo, hi) {
  const ticks = [];
  let exp = Math.floor(Math.log10(lo));
  const maxExp = Math.ceil(Math.log10(hi));
  while (exp <= maxExp) {
    const v = Math.pow(10, exp);
    if (v >= lo * 0.99 && v <= hi * 1.01) ticks.push(v);
    exp++;
  }
  if (ticks.length < 2) {
    ticks.length = 0;
    exp = Math.floor(Math.log10(lo));
    while (exp <= maxExp) {
      for (const m of [1, 2, 5]) {
        const v = m * Math.pow(10, exp);
        if (v >= lo * 0.99 && v <= hi * 1.01) ticks.push(v);
      }
      exp++;
    }
  }
  return ticks;
}

function scaleLog(v, lo, hi) {
  const logLo = Math.log10(lo), logHi = Math.log10(hi);
  if (logHi === logLo) return 0.5;
  return (Math.log10(v) - logLo) / (logHi - logLo);
}

const allDates = DATA.dates;
const xStep = allDates.length > 1 ? innerW / (allDates.length - 1) : innerW / 2;
const dateIndex = {};
allDates.forEach(function(d, i) { dateIndex[d] = i; });

function xPos(date) {
  if (allDates.length === 1) return innerW / 2;
  return dateIndex[date] * xStep;
}

DATA.families.forEach(function(fam) {
  const container = document.createElement("div");
  container.className = "chart-container";

  const title = document.createElement("div");
  title.className = "chart-title";
  title.textContent = fam.name;
  container.appendChild(title);

  let minV = Infinity, maxV = -Infinity;
  fam.series.forEach(function(s) {
    s.points.forEach(function(p) {
      if (p.value > 0) {
        if (p.value < minV) minV = p.value;
        if (p.value > maxV) maxV = p.value;
      }
    });
  });

  if (minV === Infinity) { minV = 1e-9; maxV = 1; }
  const pad = 0.15;
  const logRange = Math.log10(maxV) - Math.log10(minV);
  const lo = Math.pow(10, Math.log10(minV) - logRange * pad);
  const hi = Math.pow(10, Math.log10(maxV) + logRange * pad);

  function yPos(val) {
    return innerH - scaleLog(val, lo, hi) * innerH;
  }

  const ns = "http://www.w3.org/2000/svg";
  const svg = document.createElementNS(ns, "svg");
  svg.setAttribute("width", WIDTH);
  svg.setAttribute("height", HEIGHT);
  svg.setAttribute("viewBox", "0 0 " + WIDTH + " " + HEIGHT);

  const g = document.createElementNS(ns, "g");
  g.setAttribute("transform", "translate(" + MARGIN.left + "," + MARGIN.top + ")");
  svg.appendChild(g);

  // grid + y-axis
  const yTicks = logTicks(lo, hi);
  yTicks.forEach(function(v) {
    const y = yPos(v);
    const line = document.createElementNS(ns, "line");
    line.setAttribute("x1", 0); line.setAttribute("x2", innerW);
    line.setAttribute("y1", y); line.setAttribute("y2", y);
    line.setAttribute("stroke", "#21262d"); line.setAttribute("stroke-width", "1");
    g.appendChild(line);

    const label = document.createElementNS(ns, "text");
    label.setAttribute("x", -8); label.setAttribute("y", y + 4);
    label.setAttribute("text-anchor", "end");
    label.setAttribute("fill", "#484f58"); label.setAttribute("font-size", "11");
    label.textContent = formatValue(v);
    g.appendChild(label);
  });

  // x-axis
  const maxLabels = Math.floor(innerW / 80);
  const step = Math.max(1, Math.ceil(allDates.length / maxLabels));
  allDates.forEach(function(d, i) {
    if (i % step !== 0 && i !== allDates.length - 1) return;
    const x = xPos(d);
    const tick = document.createElementNS(ns, "line");
    tick.setAttribute("x1", x); tick.setAttribute("x2", x);
    tick.setAttribute("y1", innerH); tick.setAttribute("y2", innerH + 6);
    tick.setAttribute("stroke", "#484f58"); tick.setAttribute("stroke-width", "1");
    g.appendChild(tick);

    const parts = d.split("T");
    const label1 = document.createElementNS(ns, "text");
    label1.setAttribute("x", x); label1.setAttribute("y", innerH + 20);
    label1.setAttribute("text-anchor", "middle");
    label1.setAttribute("fill", "#484f58"); label1.setAttribute("font-size", "10");
    label1.textContent = parts[0];
    g.appendChild(label1);

    if (parts[1]) {
      const label2 = document.createElementNS(ns, "text");
      label2.setAttribute("x", x); label2.setAttribute("y", innerH + 32);
      label2.setAttribute("text-anchor", "middle");
      label2.setAttribute("fill", "#484f58"); label2.setAttribute("font-size", "10");
      label2.textContent = parts[1];
      g.appendChild(label2);
    }
  });

  // axis lines
  const xAxis = document.createElementNS(ns, "line");
  xAxis.setAttribute("x1", 0); xAxis.setAttribute("x2", innerW);
  xAxis.setAttribute("y1", innerH); xAxis.setAttribute("y2", innerH);
  xAxis.setAttribute("stroke", "#30363d"); xAxis.setAttribute("stroke-width", "1");
  g.appendChild(xAxis);

  const yAxis = document.createElementNS(ns, "line");
  yAxis.setAttribute("x1", 0); yAxis.setAttribute("x2", 0);
  yAxis.setAttribute("y1", 0); yAxis.setAttribute("y2", innerH);
  yAxis.setAttribute("stroke", "#30363d"); yAxis.setAttribute("stroke-width", "1");
  g.appendChild(yAxis);

  // series
  const seriesGroups = [];
  fam.series.forEach(function(s, si) {
    const color = COLORS[si % COLORS.length];
    const sg = document.createElementNS(ns, "g");
    sg.dataset.series = si;

    // line
    if (s.points.length > 1) {
      let d = "";
      s.points.forEach(function(p, pi) {
        const x = xPos(p.date), y = yPos(p.value);
        d += (pi === 0 ? "M" : "L") + x + "," + y;
      });
      const path = document.createElementNS(ns, "path");
      path.setAttribute("d", d);
      path.setAttribute("fill", "none");
      path.setAttribute("stroke", color);
      path.setAttribute("stroke-width", "2");
      sg.appendChild(path);
    }

    // dots
    s.points.forEach(function(p) {
      const cx = xPos(p.date), cy = yPos(p.value);
      const dot = document.createElementNS(ns, "circle");
      dot.setAttribute("cx", cx); dot.setAttribute("cy", cy);
      dot.setAttribute("r", "4");
      dot.setAttribute("fill", color);
      dot.setAttribute("stroke", "#161b22"); dot.setAttribute("stroke-width", "1.5");

      dot.addEventListener("mouseenter", function(e) {
        tooltip.style.display = "block";
        tooltip.innerHTML =
          "<span class='tooltip-label'>Size " + s.size + "</span><br>" +
          formatValue(p.value) + "<br>" +
          "<span class='tooltip-label'>" + formatDate(p.date) + "</span>";
      });
      dot.addEventListener("mousemove", function(e) {
        tooltip.style.left = (e.clientX + 12) + "px";
        tooltip.style.top = (e.clientY - 10) + "px";
      });
      dot.addEventListener("mouseleave", function() {
        tooltip.style.display = "none";
      });
      sg.appendChild(dot);
    });

    g.appendChild(sg);
    seriesGroups.push({el: sg, color: color, size: s.size, visible: true});
  });

  container.appendChild(svg);

  // legend
  const legend = document.createElement("div");
  legend.className = "legend";
  seriesGroups.forEach(function(sg, idx) {
    const item = document.createElement("div");
    item.className = "legend-item";

    const swatch = document.createElement("span");
    swatch.className = "legend-swatch";
    swatch.style.background = sg.color;

    const label = document.createElement("span");
    label.textContent = "Size " + sg.size;

    item.appendChild(swatch);
    item.appendChild(label);

    item.addEventListener("click", function() {
      sg.visible = !sg.visible;
      sg.el.style.display = sg.visible ? "" : "none";
      item.classList.toggle("hidden", !sg.visible);
    });

    legend.appendChild(item);
  });
  container.appendChild(legend);

  chartsDiv.appendChild(container);
});
</script>
</body>
</html>
`))
