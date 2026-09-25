package config

import (
	"os"
	"path/filepath"
)

type MemoryTier string

const (
	TierEco       MemoryTier = "Eco (~100MB)"
	TierBalanced  MemoryTier = "Balanced (~180MB)"
	TierStudio    MemoryTier = "Studio (~280MB)"
)

type ProfileConfig struct {
	Tier              MemoryTier
	MusicDir          string
	PreBufferSizeMB   int
	FFTWindowSize     int
	SpectrumBands     int
	MaxLibraryCache   int
	TargetSampleRate  int
	VisualizerFPS     int
}

func GetProfile(tier MemoryTier) *ProfileConfig {
	home, _ := os.UserHomeDir()
	music := filepath.Join(home, "Music")

	switch tier {
	case TierEco:
		return &ProfileConfig{
			Tier:             TierEco,
			MusicDir:         music,
			PreBufferSizeMB:  32,
			FFTWindowSize:    512,
			SpectrumBands:    24,
			MaxLibraryCache:  2000,
			TargetSampleRate: 44100,
			VisualizerFPS:    20,
		}
	case TierStudio:
		return &ProfileConfig{
			Tier:             TierStudio,
			MusicDir:         music,
			PreBufferSizeMB:  128,
			FFTWindowSize:    2048,
			SpectrumBands:    64,
			MaxLibraryCache:  20000,
			TargetSampleRate: 48000,
			VisualizerFPS:    60,
		}
	default:
		return &ProfileConfig{
			Tier:             TierBalanced,
			MusicDir:         music,
			PreBufferSizeMB:  64,
			FFTWindowSize:    1024,
			SpectrumBands:    36,
			MaxLibraryCache:  10000,
			TargetSampleRate: 44100,
			VisualizerFPS:    30,
		}
	}
}
