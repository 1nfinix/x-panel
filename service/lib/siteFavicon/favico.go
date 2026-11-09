package siteFavicon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"x-panel/lib/imagecheck"

	"github.com/PuerkitoBio/goquery"
)

// allowPrivateIPs is only set by package tests using local httptest servers.
var allowPrivateIPs bool

var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
}

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		Proxy:       nil,
		DialContext: safeDialContext,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || (req.URL.Scheme != "http" && req.URL.Scheme != "https") {
			return errors.New("redirect rejected")
		}
		return nil
	},
}

func publicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func safeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	var public []netip.Addr
	for _, ip := range ips {
		if allowPrivateIPs || publicIP(ip) {
			public = append(public, ip)
		}
	}
	if len(public) == 0 {
		return nil, errors.New("non-public address rejected")
	}
	var dialer net.Dialer
	var lastErr error
	for _, ip := range public {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

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
		iconURL, err := resp.Request.URL.Parse(href)
		if err == nil && (iconURL.Scheme == "http" || iconURL.Scheme == "https") {
			return iconURL.String(), nil
		}
	}
	return resp.Request.URL.ResolveReference(&url.URL{Path: "/favicon.ico"}).String(), nil
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
	if err != nil || (req.URL.Scheme != "http" && req.URL.Scheme != "https") || req.URL.Host == "" {
		return nil, errors.New("invalid icon URL")
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
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, errors.New("icon exceeds 1 MB")
	}
	ext, err := imagecheck.Extension(data)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(savePath, "favicon-*"+ext)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		os.Remove(file.Name())
		return nil, err
	}
	return file, nil
}

func GetOneFaviconURLAndUpload(urlStr string) (string, bool) {
	iconURL, err := GetOneFaviconURL(urlStr)
	return iconURL, err == nil
}
