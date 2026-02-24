package model

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// axisInfo describes a single metric axis.
type axisInfo struct {
	key   string
	label string
	low   string
	high  string
}

var axes = []axisInfo{
	{"formality", "Formality", "Casual", "Formal"},
	{"confidence", "Confidence", "Hedged", "Decisive"},
	{"rhythm", "Rhythm", "Uniform", "Varied"},
	{"economy", "Economy", "Expansive", "Spare"},
	{"precision", "Precision", "Vague", "Specific"},
	{"coherence", "Coherence", "Fragmented", "Structured"},
	{"vocabulary", "Vocabulary", "Plain", "Rich"},
	{"stance", "Stance", "Impersonal", "Reader-centric"},
	{"emotional_tone", "Emotional Tone", "Neutral", "Warm"},
	{"temporal_orientation", "Temporal", "Retrospective", "Prospective"},
}

// Internal messages for metrics loading.
type metricsResultMsg struct {
	Values []float64
}

type metricsErrorMsg struct {
	Err error
}

// Metrics is the metrics viewer.
type Metrics struct {
	viewport viewport.Model
	spinner  spinner.Model
	filePath string
	ctx      *ViewContext
	values   []float64
	loaded   bool
	errMsg   string
	status   string
	showHelp bool
}

const metricsHelpHeight = 1
const inkcheckInstallCmd = "go install github.com/inkcheck/inkcheck@latest"

// NewMetrics creates a new Metrics viewer for the given file.
func NewMetrics(ctx *ViewContext, filePath string) Metrics {
	vp := viewport.New(ctx.width, metricsViewportHeight(ctx, false))
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("135"))
	return Metrics{
		viewport: vp,
		spinner:  sp,
		filePath: filePath,
		ctx:      ctx,
	}
}

func (m Metrics) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		if _, err := exec.LookPath("inkcheck"); err != nil {
			return metricsErrorMsg{Err: fmt.Errorf("inkcheck not found in PATH\n\nInstall: %s", inkcheckInstallCmd)}
		}

		out, err := exec.Command("inkcheck", "signature", "-format", "json", m.filePath).Output()
		if err != nil {
			return metricsErrorMsg{Err: fmt.Errorf("inkcheck failed: %w", err)}
		}

		var result struct {
			Signature map[string]struct {
				Score float64 `json:"score"`
			} `json:"signature"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			return metricsErrorMsg{Err: fmt.Errorf("failed to parse output: %w", err)}
		}

		values := make([]float64, len(axes))
		for i, axis := range axes {
			if entry, ok := result.Signature[axis.key]; ok {
				values[i] = entry.Score
			}
		}
		return metricsResultMsg{Values: values}
	})
}

func (m Metrics) Update(msg tea.Msg) (Metrics, tea.Cmd) {
	switch msg := msg.(type) {
	case metricsResultMsg:
		m.values = msg.Values
		m.loaded = true
		m.renderContent()
		return m, nil

	case metricsErrorMsg:
		m.errMsg = msg.Err.Error()
		m.loaded = true
		m.spinner.Spinner = spinner.Pulse
		m.renderContent()
		return m, m.spinner.Tick

	case tea.WindowSizeMsg:
		m.viewport.Width = m.ctx.width
		m.viewport.Height = metricsViewportHeight(m.ctx, m.showHelp)
		if m.loaded {
			m.renderContent()
		}
		return m, nil

	case clearStatusMsg:
		m.status = ""
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "left", "h":
			if m.showHelp {
				m.showHelp = false
				m.viewport.Height = metricsViewportHeight(m.ctx, false)
				return m, nil
			}
			return m, func() tea.Msg { return CloseMetricsMsg{} }
		case "y":
			if m.errMsg != "" {
				if err := clipboard.WriteAll(inkcheckInstallCmd); err != nil {
					m.status = "Copy failed"
				} else {
					m.status = "Copied!"
				}
				return m, clearStatusAfter(2*time.Second, clearStatusMsg{})
			}
		case "?":
			m.showHelp = !m.showHelp
			m.viewport.Height = metricsViewportHeight(m.ctx, m.showHelp)
			return m, nil
		}
	}

	if !m.loaded || m.errMsg != "" {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.errMsg != "" {
			m.renderContent()
		}
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func metricsViewportHeight(ctx *ViewContext, showHelp bool) int {
	return contentHeight(ctx, metricsChromeHeight, metricsHelpHeight, showHelp)
}

func (m *Metrics) renderContent() {
	var content string
	if m.errMsg != "" {
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color("135")).
			Render(m.spinner.View() + " " + m.errMsg)
	} else {
		content = m.renderChart()
	}
	centered := centerContent(content, m.viewport.Width, m.ctx.maxWidth)
	m.viewport.SetContent(centered)
}

func (m Metrics) renderChart() string {
	barWidth := 20
	labelWidth := 0
	for _, a := range axes {
		if len(a.label) > labelWidth {
			labelWidth = len(a.label)
		}
	}

	var lines []string
	filled := lipgloss.NewStyle().Foreground(lipgloss.Color("135"))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	scoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	spectrumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Calculate max spectrum width to decide wrapping.
	spectrumWidth := 0
	for _, a := range axes {
		w := lipgloss.Width(strings.ToLower(a.low) + " ↔ " + strings.ToLower(a.high))
		if w > spectrumWidth {
			spectrumWidth = w
		}
	}
	// Full line: "  label  bar  score  spectrum"
	fullWidth := 2 + labelWidth + 2 + barWidth + 2 + 4 + 2 + spectrumWidth
	wrap := m.ctx.contentWidth() < fullWidth

	for i, axis := range axes {
		val := m.values[i]
		n := int(val*float64(barWidth) + 0.5)
		if n > barWidth {
			n = barWidth
		}
		bar := filled.Render(strings.Repeat("█", n)) +
			empty.Render(strings.Repeat("░", barWidth-n))
		label := fmt.Sprintf("%*s", labelWidth, axis.label)
		score := scoreStyle.Render(fmt.Sprintf("%.2f", val))
		spectrum := spectrumStyle.Render(
			strings.ToLower(axis.low) + " ↔ " + strings.ToLower(axis.high))
		if wrap {
			indent := strings.Repeat(" ", 2+labelWidth+2)
			lines = append(lines, fmt.Sprintf("  %s  %s  %s", label, bar, score))
			lines = append(lines, indent+spectrum)
		} else {
			lines = append(lines, fmt.Sprintf("  %s  %s  %s  %s", label, bar, score, spectrum))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Metrics) helpView() string {
	return renderHelpPane([][]helpEntry{
		{{"j/↓", "down"}, {"k/↑", "up"}, {"esc", "back"}, {"?", "help"}},
	}, m.ctx.width)
}

func (m Metrics) statusBarView() string {
	w := m.ctx.width

	left := statusBarBookName(m.ctx.bookName) + statusBarNameStyle.Render("Metrics")

	rightText := "? Help"
	if m.errMsg != "" {
		rightText = "y Copy install cmd | " + rightText
	}
	if m.status != "" {
		rightText = statusBarAccentStyle.Render(m.status) + "  " + rightText
	}
	right := statusBarHintStyle.Render(rightText)
	return statusBarFill(left, right, w)
}

func (m Metrics) View() string {
	var content string
	if !m.loaded {
		loading := m.spinner.View() + " Analysing…"
		h := metricsViewportHeight(m.ctx, false)
		content = lipgloss.Place(m.ctx.width, h, lipgloss.Center, lipgloss.Center, loading)
	} else {
		content = m.viewport.View()
	}
	var helpPane string
	if m.showHelp {
		helpPane = m.helpView()
	}
	return layoutView(logo, content, m.statusBarView(), helpPane)
}
