package markdown

import (
	"strings"
	"testing"
)


// demoTable1: 7 行 × 3 列，无 <th>，6 个单元格内含 <ul><li><p> 列表
const (
	demoTable1 = `<table style="border-collapse: collapse; width: 100%; line-height: 1.75; margin: 0 0 10px 0">
<tbody>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">大类</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">说明</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">典型例子</p></td></tr>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">指代问题</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">用户用“这个、那个、第一个、刚才那个”等表达代替具体对象</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">第二个适合生产吗？</p></li></ul><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">刚才那个项目怎么包装？</p></li></ul></td></tr>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">缺失问题</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">用户输入不完整，缺少主语、宾语、动作对象、约束条件等</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">市场占有率多少？</p></li></ul><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">帮我优化一下。</p></li></ul></td></tr>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">表达问题</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">用户表达口语化、跳跃、病句、语序混乱，或者临时改口</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">不是，我说的是另一个。</p></li></ul><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">这个不太行，换个更像面试能讲的。</p></li></ul></td></tr>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">词义问题</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">用户使用非标准术语、同义词、黑话、内部代号或多义词</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">RAG / 检索增强 / 知识库</p></li></ul><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">抓手/不够 P8</p></li></ul></td></tr>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">主观判断问题</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">用户使用依赖个人偏好、场景和标准的判断词</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">哪个最有性价比？</p></li></ul><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">再高级一点。</p></li></ul></td></tr>
<tr><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">外部事实问题</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">用户使用需要实时数据、检索结果或工具返回才能确定的表达</p></td><td style="border: 1px solid #ddd; padding: 8px 12px; vertical-align: top"><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">最近哪个框架更火？</p></li></ul><ul><li><p style="margin: 0; line-height: 1.6; white-space: pre-wrap; display: block">现在最便宜的是哪个？</p></li></ul></td></tr>
</tbody>
</table>`
)

// mdTableLines 返回 markdown 输出中所有表格行（以 | 开头的行）
func mdTableLines(md string) []string {
	var lines []string
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

// mdColumns 统计一行的列数（忽略转义的 \| ）
func mdColumns(line string) int {
	return strings.Count(strings.ReplaceAll(line, `\|`, ""), "|") - 1
}

// assertTable 校验输出包含一个结构完整的表格：
// 表头行 + 分隔行 + wantRows 个数据行，每行均为 wantCols 列。
// 返回所有表格行。
func assertTable(t *testing.T, out string, wantCols, wantRows int) []string {
	t.Helper()
	lines := mdTableLines(out)
	if len(lines) != wantRows+2 {
		t.Fatalf("want %d table lines (header+separator+%d rows), got %d:\n%s",
			wantRows+2, wantRows, len(lines), out)
	}
	if !strings.Contains(lines[1], "---") {
		t.Errorf("second line is not a separator: %q", lines[1])
	}
	for i, line := range lines {
		if c := mdColumns(line); c != wantCols {
			t.Errorf("line %d: want %d columns, got %d: %q", i, wantCols, c, line)
		}
	}
	return lines
}

func TestConvert_DemoTable1(t *testing.T) {
	out, err := getDefaultConverter().ConvertString(demoTable1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("demoTable1 output:\n%s", out)

	// 无 <th> 的表格，首行应被提升为表头（WithHeaderPromotion）
	lines := assertTable(t, out, 3, 6)
	for _, want := range []string{"大类", "说明", "典型例子"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("header row missing %q: %q", want, lines[0])
		}
	}

	// 单元格内的 <ul> 列表不应导致整张表被放弃转换，
	// 列表项内容应保留在单元格内（以 <br /> 分隔）
	for _, want := range []string{"第二个适合生产吗？", "刚才那个项目怎么包装？", "市场占有率多少？"} {
		if !strings.Contains(out, want) {
			t.Errorf("list item %q lost from output:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "<br />") {
		t.Errorf("list items in cell should be joined with <br />:\n%s", out)
	}

	assertNoResidualTags(t, out)
}

// assertNoResidualTags 确认 HTML 标签没有残留到 markdown 输出中
func assertNoResidualTags(t *testing.T, out string) {
	t.Helper()
	for _, tag := range []string{"<td", "<th", "<tr", "<ul", "<ol", "<li", "<table", "</table"} {
		if strings.Contains(out, tag) {
			t.Errorf("output contains residual html tag %q:\n%s", tag, out)
		}
	}
}

func TestConvert_PipeInTableCell(t *testing.T) {
	html := `<table><tr><td>a|b</td><td>c</td></tr><tr><td>x</td><td>y</td></tr></table>`
	out, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `a\|b`) {
		t.Errorf("pipe in cell should be escaped as \\|:\n%s", out)
	}
	assertTable(t, out, 2, 1)
}

func TestConvert_ColspanTableCell(t *testing.T) {
	html := `<table>` +
		`<tr><td colspan="2">wide</td><td>c</td></tr>` +
		`<tr><td>a</td><td>b</td><td>c</td></tr>` +
		`</table>`
	out, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}
	lines := assertTable(t, out, 3, 1)
	if !strings.Contains(lines[0], "wide") {
		t.Errorf("header row missing spanned cell content: %q", lines[0])
	}
}

func TestConvert_RowspanTableCell(t *testing.T) {
	html := `<table>` +
		`<tr><td rowspan="2">r</td><td>a</td></tr>` +
		`<tr><td>b</td></tr>` +
		`</table>`
	out, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}
	lines := assertTable(t, out, 2, 1)
	if !strings.Contains(lines[0], "r") {
		t.Errorf("header row missing rowspan cell content: %q", lines[0])
	}
}

func TestConvert_MultipleParagraphsInCell(t *testing.T) {
	html := `<table><tr><td><p>line1</p><p>line2</p></td><td>x</td></tr></table>`
	out, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<br />") {
		t.Errorf("multiple paragraphs in cell should be joined with <br />:\n%s", out)
	}
	assertTable(t, out, 2, 0)
}

func TestConvert_ListOutsideTableUnaffected(t *testing.T) {
	html := `<p>before</p><ul><li>item1</li><li>item2</li></ul><p>after</p>`
	out, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- item1") || !strings.Contains(out, "- item2") {
		t.Errorf("list outside table should stay a markdown list:\n%s", out)
	}
}
