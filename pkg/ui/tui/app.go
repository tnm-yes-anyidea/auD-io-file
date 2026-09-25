package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tnm-yes-anyidea/auD-io-file/pkg/audio"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/config"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/library"
)

type ViewTab int
const (
	TabPlayer ViewTab = iota
	TabEQ
	TabVisualizer
)

var (
	cyanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	magentaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

type navItem struct{ title, desc, targetType string }
func (n navItem) Title() string       { return n.title }
func (n navItem) Description() string { return n.desc }
func (n navItem) FilterValue() string { return n.title }

type trackItem struct{ track library.Track }
func (t trackItem) Title() string       { return fmt.Sprintf("%s - %s", t.track.Artist, t.track.Title) }
func (t trackItem) Description() string { return fmt.Sprintf("[%s] %s | %s", t.track.Album, t.track.Genre, filepath.Ext(t.track.Path)) }
func (t trackItem) FilterValue() string { return fmt.Sprintf("%s %s %s %s", t.track.Title, t.track.Artist, t.track.Album, t.track.Genre) }

type listState struct {
	title string
	items []list.Item
}

type tickMsg time.Time

type Model struct {
	cfg          *config.ProfileConfig
	engine       *audio.Engine
	lib          *library.Library
	currentTab   ViewTab
	list         list.Model
	navStack     []listState
	eqGains      [5]float64
	activeEQ     int
	width        int
	height       int
	artRender    string
	currentHover string
}

func NewModel(cfg *config.ProfileConfig, engine *audio.Engine) *Model {
	return &Model{
		cfg:        cfg,
		engine:     engine,
		currentTab: TabPlayer,
		navStack:   make([]listState, 0),
	}
}

func tickCmd(fps int) tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(m.cfg.VisualizerFPS),
		func() tea.Msg {
			lib, _ := library.ScanLibrary(m.cfg.MusicDir, m.cfg.MaxLibraryCache)
			return lib
		},
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Guard against negative/zero dimensions when scaling terminal
		listWidth := (msg.Width / 2) - 4
		if listWidth < 0 { listWidth = 0 }
		listHeight := msg.Height - 12
		if listHeight < 0 { listHeight = 0 }

		if m.list.Title == "" {
			delegate := list.NewDefaultDelegate()
			m.list = list.New([]list.Item{}, delegate, listWidth, listHeight)
			m.list.Title = "Loading Library..."
		} else {
			m.list.SetSize(listWidth, listHeight)
		}

	case *library.Library:
		m.lib = msg
		rootItems := []list.Item{
			navItem{title: "All Tracks", desc: fmt.Sprintf("%d Tracks", len(msg.Tracks)), targetType: "all"},
			navItem{title: "Albums", desc: fmt.Sprintf("%d Albums", len(msg.Albums)), targetType: "albums"},
			navItem{title: "Artists", desc: fmt.Sprintf("%d Artists", len(msg.Artists)), targetType: "artists"},
		}
		m.list.Title = "Library Menu"
		m.list.SetItems(rootItems)

	case tickMsg:
		return m, tickCmd(m.cfg.VisualizerFPS)
		
	case tea.MouseMsg:
		// Safe mouse click handling
		if msg.Type == tea.MouseLeft && msg.Y > m.height-4 && m.width > 0 {
			length := m.engine.Length()
			if length > 0 {
				pct := float64(msg.X) / float64(m.width)
				target := time.Duration(float64(length) * pct)
				m.engine.Seek(target)
			}
		}

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering { break }

		switch msg.String() {
		case "q": return m, tea.Quit
		case "1": m.currentTab = TabPlayer
		case "2": m.currentTab = TabEQ
		case "3": m.currentTab = TabVisualizer
		
		case "c", " ": m.engine.TogglePause()
		case "v": m.engine.Stop()
		case "-", "_": m.engine.SetVolume(-0.5)
		case "=", "+": m.engine.SetVolume(0.5)
		
		case "r":
			m.engine.Repeat = (m.engine.Repeat + 1) % 3
			
		case "left":
			m.engine.SeekRelative(-time.Second * 5)
		case "right":
			m.engine.SeekRelative(time.Second * 5)

		case "b":
			if m.engine.QueueIndex+1 < len(m.engine.Queue) {
				_ = m.engine.PlayQueueIndex(m.engine.QueueIndex + 1)
			}
		case "z":
			if m.engine.QueueIndex > 0 {
				_ = m.engine.PlayQueueIndex(m.engine.QueueIndex - 1)
			}

		case "backspace", "esc":
			if m.currentTab == TabPlayer && len(m.navStack) > 0 {
				last := m.navStack[len(m.navStack)-1]
				m.navStack = m.navStack[:len(m.navStack)-1]
				m.list.Title = last.title
				m.list.SetItems(last.items)
			}

		case "enter":
			if m.currentTab == TabPlayer {
				selected := m.list.SelectedItem()
				if nav, ok := selected.(navItem); ok {
					m.navStack = append(m.navStack, listState{title: m.list.Title, items: m.list.Items()})
					var newItems []list.Item
					m.list.Title = nav.title
					switch nav.targetType {
					case "all":          for _, t := range m.lib.Tracks { newItems = append(newItems, trackItem{track: t}) }
					case "albums":       for k := range m.lib.Albums { newItems = append(newItems, navItem{title: k, desc: "Album", targetType: "group_album"}) }
					case "artists":      for k := range m.lib.Artists { newItems = append(newItems, navItem{title: k, desc: "Artist", targetType: "group_artist"}) }
					case "group_album":  for _, t := range m.lib.Albums[nav.title] { newItems = append(newItems, trackItem{track: t}) }
					case "group_artist": for _, t := range m.lib.Artists[nav.title] { newItems = append(newItems, trackItem{track: t}) }
					}
					m.list.SetItems(newItems)
				} else if trk, ok := selected.(trackItem); ok {
					m.engine.Queue = make([]string, 0)
					playIdx := 0
					for i, itm := range m.list.Items() {
						if ti, ok := itm.(trackItem); ok {
							m.engine.Queue = append(m.engine.Queue, ti.track.Path)
							if ti.track.Path == trk.track.Path { playIdx = i }
						}
					}
					_ = m.engine.PlayQueueIndex(playIdx)
					m.artRender = RenderAlbumArt(library.ExtractArt(trk.track.Path), 30, 15)
				}
			}
		
		case "h": if m.currentTab == TabEQ && m.activeEQ > 0 { m.activeEQ-- }
		case "l": if m.currentTab == TabEQ && m.activeEQ < 4 { m.activeEQ++ }
		case "k":
			if m.currentTab == TabEQ && m.eqGains[m.activeEQ] < 12.0 {
				m.eqGains[m.activeEQ] += 1.0
				m.engine.SetEQGain(m.activeEQ, m.eqGains[m.activeEQ])
			}
		case "j":
			if m.currentTab == TabEQ && m.eqGains[m.activeEQ] > -12.0 {
				m.eqGains[m.activeEQ] -= 1.0
				m.engine.SetEQGain(m.activeEQ, m.eqGains[m.activeEQ])
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	
	if sel, ok := m.list.SelectedItem().(trackItem); ok {
		if m.currentHover != sel.track.Path {
			m.currentHover = sel.track.Path
			m.artRender = RenderAlbumArt(library.ExtractArt(sel.track.Path), 30, 15)
		}
	} else {
		m.currentHover = ""
		m.artRender = emptyArtBox(30, 15)
	}

	return m, cmd
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	return fmt.Sprintf("%02d:%02d", m, s)
}

func (m *Model) View() string {
	// Guard against uninitialized or extreme tiny windows
	if m.width <= 0 || m.height <= 0 {
		return "Initializing UI..."
	}

	var b strings.Builder
	tabs := []string{"[1] Player", "[2] EQ", "[3] Scopes"}
	var renderedTabs []string
	for i, t := range tabs {
		if ViewTab(i) == m.currentTab {
			renderedTabs = append(renderedTabs, cyanStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, subtleStyle.Render(t))
		}
	}
	b.WriteString(strings.Join(renderedTabs, "  |  ") + "\n\n")

	switch m.currentTab {
	case TabPlayer:
		listStr := m.list.View()
		artStr := boxStyle.Render(m.artRender + "\n" + cyanStyle.Render("Now Playing: ") + filepath.Base(m.engine.CurrentTrack))
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listStr, "   ", artStr))

	case TabEQ:
		b.WriteString(magentaStyle.Render("=== 5-Band Parametric Equalizer ===\n\n"))
		freqs := []string{"60Hz", "250Hz", "1kHz", "4kHz", "12kHz"}
		for i := 0; i < 5; i++ {
			sel := "  "
			if i == m.activeEQ { sel = ">>" }
			b.WriteString(fmt.Sprintf("%s %-6s [%+5.1f dB]  %s\n", sel, freqs[i], m.eqGains[i], renderSlider(m.eqGains[i], -12, 12, 24)))
		}

	case TabVisualizer:
		b.WriteString(cyanStyle.Render("=== Scope Rack ===\n\n"))
		spec, scopeL, scopeR := m.engine.Visualizer.Snapshot()
		b.WriteString("Spectrum:\n" + renderSpectrumBars(spec, 12) + "\n\n")
		b.WriteString("Vectorscope:\n" + renderVectorscope(scopeL, scopeR, 14, 28) + "\n")
	}

	b.WriteString("\n" + strings.Repeat("─", m.width) + "\n")
	
	// Progress Bar Render Guards
	pos := m.engine.Position()
	ln := m.engine.Length()
	barWidth := m.width - 30
	if barWidth < 0 { barWidth = 0 }
	barStr := renderProgressBar(pos, ln, barWidth)
	
	state := "⏹ STOPPED"
	if m.engine.IsPlaying { state = "▶ PLAYING" } else if m.engine.CurrentTrack != "" { state = "⏸ PAUSED" }
	
	repStr := "Off"
	if m.engine.Repeat == audio.RepeatAll { repStr = "All" } else if m.engine.Repeat == audio.RepeatOne { repStr = "One" }

	statusStr := fmt.Sprintf("%s | Rep:%s | Vol:%+.1fdB\n%s %s / %s", 
		state, repStr, m.engine.VolumeLevel, barStr, formatDuration(pos), formatDuration(ln))
		
	b.WriteString(boxStyle.Render(statusStr))

	return b.String()
}

// ---------------------------------------------------------
// SAFE RENDERING HELPERS
// ---------------------------------------------------------

func renderProgressBar(pos, length time.Duration, width int) string {
	if width <= 0 { return "" }
	if length <= 0 { return strings.Repeat("-", width) }
	
	pct := float64(pos) / float64(length)
	filled := int(pct * float64(width))
	if filled < 0 { filled = 0 }
	if filled > width { filled = width }
	
	var sb strings.Builder
	sb.WriteString(strings.Repeat("=", filled))
	
	if filled < width {
		sb.WriteString(">")
		rem := width - filled - 1
		if rem > 0 {
			sb.WriteString(strings.Repeat("-", rem))
		}
	}
	return sb.String()
}

func renderSlider(val, min, max float64, width int) string {
	if width <= 0 { return "" }
	pos := int(((val - min) / (max - min)) * float64(width))
	if pos < 0 { pos = 0 }
	if pos >= width { pos = width - 1 }

	var s strings.Builder
	s.WriteString("[")
	for i := 0; i < width; i++ {
		if i == pos { s.WriteString("█") } else if i == width/2 { s.WriteString("│") } else { s.WriteString("─") }
	}
	return s.String() + "]"
}

func renderSpectrumBars(bands []float64, height int) string {
	if height <= 0 { return "" }
	blocks := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var lines []string
	
	for h := height; h > 0; h-- {
		var line strings.Builder
		for _, b := range bands {
			scaled := int(b * float64(height) * 3)
			if scaled >= h { line.WriteString("█ ") } else if scaled == h-1 { line.WriteString(blocks[4] + " ") } else { line.WriteString("  ") }
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

func renderVectorscope(left, right []float64, rows, cols int) string {
	if rows <= 0 || cols <= 0 { return "" }
	grid := make([][]rune, rows)
	for r := range grid {
		grid[r] = make([]rune, cols)
		for c := range grid[r] { grid[r][c] = ' ' }
	}
	
	midR, midC := rows/2, cols/2
	points := len(left)
	if len(right) < points { points = len(right) }
	
	for i := 0; i < points; i++ {
		r := midR - int(((left[i]+right[i])*0.707)*float64(midR)*0.9)
		c := midC + int(((left[i]-right[i])*0.707)*float64(midC)*0.9)
		
		if r >= 0 && r < rows && c >= 0 && c < cols { 
			grid[r][c] = '•' 
		}
	}
	
	var sb strings.Builder
	for r := 0; r < rows; r++ { sb.WriteString("│" + string(grid[r]) + "│\n") }
	return sb.String()
}
