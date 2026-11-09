package wallpaper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"x-panel/global"
	"time"
)

const feedURL = "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=8"
const maxImages = 30

type imageInfo struct {
	URL     string `json:"url"`
	URLBase string `json:"urlbase"`
}

type feed struct {
	Images []imageInfo `json:"images"`
}

func Start(sourcePath string) {
	go func() {
		client := &http.Client{Timeout: 10 * time.Second}
		run := func() {
			if err := syncFromURL(context.Background(), client, feedURL, sourcePath); err != nil {
				global.Logger.Warnf("wallpaper sync: %v", err)
			}
		}
		run()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}

func syncFromURL(ctx context.Context, client *http.Client, feedAddress, sourcePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedAddress, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feed returned HTTP %d", resp.StatusCode)
	}
	var catalog feed
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&catalog); err != nil {
		return err
	}
	dir := filepath.Join(sourcePath, "wallpaper")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	feedBase, err := url.Parse(feedAddress)
	if err != nil {
		return err
	}
	var failures []error
	for _, image := range catalog.Images {
		imageURL, err := feedBase.Parse(image.URL)
		if err != nil || (imageURL.Scheme != "https" && imageURL.Scheme != "http") {
			failures = append(failures, errors.New("invalid wallpaper URL"))
			continue
		}
		key := image.URLBase
		if key == "" {
			key = imageURL.Path
		}
		hash := sha256.Sum256([]byte(key))
		ext := strings.ToLower(path.Ext(imageURL.Path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			ext = ".jpg"
		}
		filename := hex.EncodeToString(hash[:]) + ext
		target := filepath.Join(dir, filename)
		if info, err := os.Stat(target); err == nil && info.Size() > 0 {
			continue
		}
		if err := download(ctx, client, imageURL.String(), target); err != nil {
			failures = append(failures, err)
		}
	}
	if err := prune(dir, maxImages); err != nil {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}

func download(ctx context.Context, client *http.Client, address, target string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wallpaper returned HTTP %d", resp.StatusCode)
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".wallpaper-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	n, err := io.Copy(file, io.LimitReader(resp.Body, 10*1024*1024+1))
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if n > 10*1024*1024 {
		return errors.New("wallpaper exceeds 10 MB")
	}
	return os.Rename(file.Name(), target)
}

func prune(dir string, limit int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	type stored struct {
		name string
		time time.Time
	}
	files := make([]stored, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		files = append(files, stored{entry.Name(), info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].time.Equal(files[j].time) {
			return files[i].name < files[j].name
		}
		return files[i].time.Before(files[j].time)
	})
	excess := len(files) - limit
	if excess < 0 {
		excess = 0
	}
	for _, file := range files[:excess] {
		if err := os.Remove(filepath.Join(dir, file.name)); err != nil {
			return err
		}
	}
	return nil
}

func PublicURLs(sourcePath string) ([]string, error) {
	if !strings.HasPrefix(sourcePath, "./") {
		return nil, errors.New("source_path must start with ./")
	}
	entries, err := os.ReadDir(filepath.Join(sourcePath, "wallpaper"))
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(entries))
	prefix := "/" + filepath.ToSlash(strings.TrimPrefix(sourcePath, "./")) + "/wallpaper/"
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			urls = append(urls, prefix+entry.Name())
		}
	}
	sort.Strings(urls)
	return urls, nil
}
