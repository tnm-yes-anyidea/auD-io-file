package gui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tnm-yes-anyidea/auD-io-file/pkg/audio"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/config"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/library"
)

func LaunchGUI(cfg *config.ProfileConfig, engine *audio.Engine) {
	// Initialize the Fyne application and window
	a := app.New()
	w := a.NewWindow("auD-io-file - Standalone Desktop UI")
	w.Resize(fyne.NewSize(900, 600))

	// Scan the library
	lib, _ := library.ScanLibrary(cfg.MusicDir, cfg.MaxLibraryCache)

	// Setup Cover Art Canvas (starts empty)
	coverArt := canvas.NewImageFromResource(nil)
	coverArt.FillMode = canvas.ImageFillContain
	coverArt.SetMinSize(fyne.NewSize(350, 350))

	titleLabel := widget.NewLabel("Select a track from the library to play")
	titleLabel.Alignment = fyne.TextAlignCenter
	
	progress := widget.NewProgressBar()

	// Extract display titles for the list
	var tracks []string
	for _, t := range lib.Tracks {
		display := t.Artist + " - " + t.Title
		if display == "Unknown - Unknown" {
			display = t.Path
		}
		tracks = append(tracks, display)
	}
	
	// Create the scrollable list of tracks
	trackList := widget.NewList(
		func() int { return len(tracks) },
		func() fyne.CanvasObject { return widget.NewLabel("Template Track Name That Is Long Enough") },
		func(i widget.ListItemID, o fyne.CanvasObject) { 
			o.(*widget.Label).SetText(tracks[i]) 
		},
	)

	// Handle track selection (Mouse Click)
	trackList.OnSelected = func(id widget.ListItemID) {
		// Guard against out of bounds
		if int(id) < 0 || int(id) >= len(lib.Tracks) {
			return
		}
		
		trk := lib.Tracks[id]
		
		// Rebuild the engine queue so auto-play next track works
		engine.Queue = make([]string, 0)
		for _, t := range lib.Tracks {
			engine.Queue = append(engine.Queue, t.Path)
		}
		
		// Play selected track
		_ = engine.PlayQueueIndex(int(id))
		titleLabel.SetText(trk.Artist + " - " + trk.Title)
		
		// Extract native high-res image and render it
		imgData := library.ExtractArt(trk.Path)
		if len(imgData) > 0 {
			res := fyne.NewStaticResource("cover.jpg", imgData)
			coverArt.Resource = res
		} else {
			coverArt.Resource = nil
		}
		coverArt.Refresh()
	}

	// Playback Controls
	playBtn := widget.NewButton("Play / Pause", func() { engine.TogglePause() })
	stopBtn := widget.NewButton("Stop", func() { engine.Stop() })
	
	controls := container.NewGridWithColumns(2, playBtn, stopBtn)
	
	// Background routine to update the progress bar seamlessly
	go func() {
		for range time.Tick(time.Second) {
			if engine.IsPlaying {
				ln := engine.Length()
				if ln > 0 {
					progress.SetValue(float64(engine.Position()) / float64(ln))
				}
			} else if engine.CurrentTrack == "" {
				// Reset bar if stopped
				progress.SetValue(0)
			}
		}
	}()

	// Build the Layout
	leftPane := container.NewBorder(nil, nil, nil, nil, trackList)
	rightPane := container.NewVBox(
		coverArt,
		titleLabel,
		progress,
		controls,
	)

	// Split the screen 40% library / 60% player
	split := container.NewHSplit(leftPane, rightPane)
	split.Offset = 0.4

	w.SetContent(split)
	w.ShowAndRun()
}
