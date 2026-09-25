# auD-io-file 🎵

An ultra-lightweight, audiophile-grade music player written in Golang. 

Designed as a high-performance alternative to Electron-based players, `auD-io-file` operates strictly within a **100MB–300MB RAM budget**. It features a custom DSP engine, real-time visualizers, and dual interfaces: a hyper-fast `cmus`-style Terminal UI (TUI) and a cross-platform Desktop GUI.

---

## 🧠 AI / Developer Context (System Prompt)

**Architecture & Stack:**
- **Language:** Go 1.21+
- **Audio Engine:** `github.com/gopxl/beep` (Direct ALSA/CoreAudio/WASAPI hardware interfacing)
- **TUI Framework:** `github.com/charmbracelet/bubbletea` & `bubbles/list` (ANSI album art, mouse events)
- **GUI Framework:** `fyne.io/fyne/v2` (Native desktop windows, true image rendering)
- **Metadata:** `github.com/dhowden/tag` (ID3v2, Vorbis tags, embedded art extraction)
- **DSP Pipeline:** Real-time Cooley-Tukey Radix-2 FFT (Spectrum), Lissajous Vectorscope, and Robert Bristow-Johnson (RBJ) Biquad Parametric EQ (5-band).
- **Memory Management:** Configurable pre-allocation ring buffers and circular arrays to prevent GC spikes and control RAM usage.

**Design Philosophy:**
Zero-bloat, modular, and gapless-capable. Audio decoding is streamed natively without buffering entire files into RAM. UI and Engine are decoupled, allowing seamless switching between the terminal and standalone graphical application via CLI flags.

---

## ✨ Features

- **Dual Interfaces:** Run as a TUI (`-ui=tui`) with ANSI-rendered album art, or as a native Desktop GUI (`-ui=gui`) with high-res graphics.
- **Audiophile DSP Engine:** Built-in 5-Band Parametric Equalizer, FFT Spectrum Analyzer, and Stereo Phase Vectorscope.
- **Memory Governors:** Dial in your RAM usage with `-tier=eco` (~100MB), `balanced` (~180MB), or `studio` (~280MB @ 60FPS visualizers).
- **MOCP/Cmus Workflow:** Lightning-fast keyboard navigation, instant search, drill-down library browser, and gapless queue management.
- **Dynamic Resampling:** Automatically resamples tracks to match your hardware DAC's target sample rate on the fly.

---

## 🛠 Installation & Prerequisites

Because this player interfaces directly with your hardware audio mixer and native window managers, you need a few C libraries installed on your system.

### 1. Install System Dependencies
**Arch / Manjaro Linux:**
```bash
sudo pacman -S base-devel pkgconf alsa-lib xorg-server-devel libxcursor libxrandr libxinerama libxi

```

**Ubuntu / Debian:**

```bash
sudo apt install pkg-config libasound2-dev libgl1-mesa-dev xorg-dev

```

**Fedora:**

```bash
sudo dnf install pkgconf-pkg-config alsa-lib-devel mesa-libGL-devel

```

### 2. Download and Build

```bash
git clone [https://github.com/tnm-yes-anyidea/auD-io-file.git](https://github.com/tnm-yes-anyidea/auD-io-file.git)
cd auD-io-file
go mod tidy
make build

```

---

## 🚀 Usage

Launch the player using the `make run` command, or directly via Go to pass specific execution flags.

**Start the Terminal UI (Default):**

```bash
go run ./cmd/app

```

**Start the Native Desktop GUI:**

```bash
go run ./cmd/app -ui=gui

```

**Set Memory / Performance Tier:**

```bash
go run ./cmd/app -tier=eco     # Lowest RAM footprint (~100MB), 20FPS visualizers
go run ./cmd/app -tier=studio  # Highest fidelity (~280MB), 60FPS visualizers, massive cache

```

---

## ⌨️ TUI Keybindings (cmus/mocp style)

**Global / Playback Controls:**

* `c` or `Space` : Play / Pause
* `v` : Stop
* `z` : Previous Track
* `b` : Next Track
* `r` : Toggle Repeat Mode (Off / All / One)
* `-` / `_` : Decrease Volume
* `=` / `+` : Increase Volume
* `Left Arrow` : Seek Backward (5s)
* `Right Arrow` : Seek Forward (5s)
* `q` : Quit Application

**Navigation & Library:**

* `1`, `2`, `3` : Switch Tabs (Player, EQ, Visualizers)
* `Up` / `Down` : Navigate lists
* `Enter` : Dive into Album/Artist/Genre or Play Track
* `Backspace` / `Esc` : Go back / Up one level
* `/` : Open fuzzy-finder search (searches Title, Artist, Album, Genre simultaneously)
* *Mouse Support:* You can click the progress bar at the bottom to seek to a specific time.

**Parametric EQ (Tab 2):**

* `h` / `l` : Select EQ Band (60Hz, 250Hz, 1kHz, 4kHz, 12kHz)
* `k` / `j` : Increase / Decrease Gain (+/- 12dB)

---

## 📂 Project Structure

```text
auD-io-file/
├── Makefile                # Build scripts
├── cmd/
│   └── app/
│       └── main.go         # Application entry point and flag parsing
└── pkg/
    ├── audio/
    │   └── engine.go       # Audio routing, playback state, queue, and volume
    ├── config/
    │   └── memory.go       # Memory tiers and configuration variables
    ├── dsp/
    │   ├── biquad.go       # RBJ Parametric EQ math
    │   ├── fft_visualizer.go # Cooley-Tukey FFT and Vectorscope processing
    │   └── ringbuffer.go   # Lock-free circular memory allocation
    ├── library/
    │   └── scanner.go      # ID3/Vorbis parsing and hierarchy mapping
    └── ui/
        ├── tui/
        │   ├── app.go      # BubbleTea terminal interface and input routing
        │   └── art.go      # ANSI TrueColor image downscaler
        └── gui/
            └── app.go      # Fyne native window implementation

```
<video src="bin/output.gif" controls="controls" muted="muted" width="100%"></video>
