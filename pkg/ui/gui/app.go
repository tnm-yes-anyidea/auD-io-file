package gui

import (
	"bytes"
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
	a := app.New()
	w := a.NewWindow("auD-io-file - Standalone UI")
	w.Resize(fyne.NewSize(800, 600))

	lib, _ := library.ScanLibrary(cfg.MusicDir, cfg.MaxLibraryCache)

	// UI Components
	coverArt := canvas.NewImageFromResource(nil)
	coverArt.FillMode = canvas.ImageFillContain
	coverArt.SetMinSize(fyne.NewSize(300, 300))

	titleLabel := widget.NewLabel("Select a track to play")
	progress := widget.NewProgressBar()

	// Track List
	var tracks []string
	for _, t := range lib.Tracks { tracks = append(tracks, t.Title) }
	
	trackList := widget.NewList(
		func() int { return len(tracks) },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(tracks[i]) },
	)

	trackList.OnSelected = func(id widget.ListItemID) {
		trk := lib.Tracks[id]
		_ = engine.Play(trk.Path)
		titleLabel.SetText(trk.Artist + " - " + trk.Title)
		
		// Render Native Image
		imgData := library.ExtractArt(trk.Path)
		if len(imgData) > 0 {
			res := fyne.NewStaticResource("cover.jpg", imgData)
			coverArt.Resource = res
		} else {
			coverArt.Resource = nil
		}
		coverArt.Refresh()
	}

	// Controls
	playBtn := widget.NewButton("Play / Pause", func() { engine.TogglePause() })
	stopBtn := widget.NewButton("Stop", func() { engine.Stop() })
	
	controls := container.NewHBox(playBtn, stopBtn)
	
	// Update UI Loop
	go func() {
		for range time.Tick(time.Second) {
			if engine.IsPlaying {
				ln := engine.Length()
				if ln > 0 {
					progress.SetValue(float64(engine.Position()) / float64(ln))
				}
			}
		}
	}()

	leftPane := container.NewBorder(nil, nil, nil, nil, trackList)
	rightPane := container.NewVBox(
		coverArt,
		titleLabel,
		progress,
		controls,
	)

	split := container.NewHSplit(leftPane, rightPane)
	split.Offset = 0.4

	w.SetContent(split)
	w.ShowAndRun()
}
