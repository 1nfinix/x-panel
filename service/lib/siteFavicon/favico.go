package siteFavicon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var client = &http.Client{Timeout: 10 * time.Second}

func GetOneFaviconURL(urlStr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return GetOneFaviconURLContext(ctx, urlStr)
}

func GetOneFaviconURLContext(ctx context.Context, urlStr string) (string, error) {
	pageURL, err := url.Parse(urlStr)
	if err != nil || (pageURL.Scheme != "http" && pageURL.Scheme != "https") || pageURL.Host == "" {
		return "", errors.New("invalid website URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("website returned HTTP %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", err
	}
	if href := selectFavicon(doc); href != "" {
		iconURL, err := pageURL.Parse(href)
		if err == nil && (iconURL.Scheme == "http" || iconURL.Scheme == "https") {
			return iconURL.String(), nil
		}
	}
	return pageURL.ResolveReference(&url.URL{Path: "/favicon.ico"}).String(), nil
}

func selectFavicon(doc *goquery.Document) string {
	bestHref, bestSize, bestApple := "", -1, false
	doc.Find("link[rel][href]").Each(func(_ int, s *goquery.Selection) {
		rel := strings.ToLower(s.AttrOr("rel", ""))
		if !strings.Contains(rel, "icon") {
			return
		}
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" {
			return
		}
		size := iconSize(s.AttrOr("sizes", ""))
		apple := strings.Contains(rel, "apple-touch-icon")
		if size > bestSize || (size == bestSize && apple && !bestApple) {
			bestHref, bestSize, bestApple = href, size, apple
		}
	})
	return bestHref
}

func iconSize(sizes string) int {
	maxSize := 0
	for _, token := range strings.Fields(strings.ToLower(sizes)) {
		parts := strings.Split(token, "x")
		if len(parts) != 2 {
			continue
		}
		width, widthErr := strconv.Atoi(parts[0])
		height, heightErr := strconv.Atoi(parts[1])
		if widthErr == nil && heightErr == nil && width > 0 && height > 0 && width*height > maxSize {
			maxSize = width * height
		}
	}
	return maxSize
}

func DownloadImage(urlStr, savePath string, maxSize int64) (*os.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return DownloadImageContext(ctx, urlStr, savePath, maxSize)
}

func DownloadImageContext(ctx context.Context, urlStr, savePath string, maxSize int64) (*os.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("icon returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxSize {
		return nil, errors.New("icon exceeds 1 MB")
	}
	ext := path.Ext(req.URL.Path)
	if ext == "" || len(ext) > 8 {
		ext = ".ico"
	}
	file, err := os.CreateTemp(savePath, "favicon-*"+ext)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	n, copyErr := io.Copy(file, io.LimitReader(resp.Body, maxSize+1))
	if copyErr != nil || n > maxSize {
		os.Remove(file.Name())
		if copyErr != nil {
			return nil, copyErr
		}
		return nil, errors.New("icon exceeds 1 MB")
	}
	return file, nil
}

func GetOneFaviconURLAndUpload(urlStr string) (string, bool) {
	iconURL, err := GetOneFaviconURL(urlStr)
	return iconURL, err == nil
}
