package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tnm-yes-anyidea/auD-io-file/pkg/audio"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/config"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/ui/gui"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/ui/tui"
)

func main() {
	uiFlag := flag.String("ui", "tui", "Select interface: 'tui' (Terminal) or 'gui' (Standalone App)")
	tierFlag := flag.String("tier", "balanced", "Memory profile: eco, balanced, studio")
	flag.Parse()

	var selectedTier config.MemoryTier
	switch *tierFlag {
	case "eco":    selectedTier = config.TierEco
	case "studio": selectedTier = config.TierStudio
	default:       selectedTier = config.TierBalanced
	}

	cfg := config.GetProfile(selectedTier)

	engine, err := audio.NewEngine(cfg)
	if err != nil {
		fmt.Printf("Hardware Audio Init Failed: %v\n", err)
		os.Exit(1)
	}

	if *uiFlag == "gui" {
		// Launch cross-platform Desktop GUI
		gui.LaunchGUI(cfg, engine)
	} else {
		// Launch BubbleTea Terminal UI (Mouse Enabled)
		app := tui.NewModel(cfg, engine)
		p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

		if _, err := p.Run(); err != nil {
			fmt.Printf("Fatal TUI Loop Error: %v\n", err)
			os.Exit(1)
		}
	}
}
