package library

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

type Track struct {
	Path   string
	Title  string
	Artist string
	Album  string
	Genre  string
}

type Library struct {
	Tracks  []Track
	Artists map[string][]Track
	Albums  map[string][]Track
	Genres  map[string][]Track
}

func ScanLibrary(dir string, maxItems int) (*Library, error) {
	lib := &Library{
		Tracks:  make([]Track, 0),
		Artists: make(map[string][]Track),
		Albums:  make(map[string][]Track),
		Genres:  make(map[string][]Track),
	}

	valid := map[string]bool{".flac": true, ".mp3": true, ".ogg": true, ".wav": true}

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if len(lib.Tracks) >= maxItems {
			return filepath.SkipDir
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !valid[ext] {
			return nil
		}

		t := Track{Path: path, Title: filepath.Base(path), Artist: "Unknown", Album: "Unknown", Genre: "Unknown"}
		f, err := os.Open(path)
		if err == nil {
			if m, err := tag.ReadFrom(f); err == nil {
				if m.Title() != "" { t.Title = m.Title() }
				if m.Artist() != "" { t.Artist = m.Artist() }
				if m.Album() != "" { t.Album = m.Album() }
				if m.Genre() != "" { t.Genre = m.Genre() }
			}
			f.Close()
		}

		lib.Tracks = append(lib.Tracks, t)
		lib.Artists[t.Artist] = append(lib.Artists[t.Artist], t)
		lib.Albums[t.Album] = append(lib.Albums[t.Album], t)
		lib.Genres[t.Genre] = append(lib.Genres[t.Genre], t)

		return nil
	})

	return lib, nil
}

func ExtractArt(filePath string) []byte {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil || m.Picture() == nil {
		return nil
	}
	return m.Picture().Data
}
