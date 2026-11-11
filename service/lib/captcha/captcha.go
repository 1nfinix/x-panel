package captcha

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type captchaEntry struct {
	answer  string
	expires time.Time
}

type ExpiringStore struct {
	mu      sync.Mutex
	entries map[string]captchaEntry
	now     func() time.Time
}

func NewExpiringStore() *ExpiringStore {
	return &ExpiringStore{entries: make(map[string]captchaEntry), now: time.Now}
}

func (s *ExpiringStore) Set(id, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, entry := range s.entries {
		if !s.now().Before(entry.expires) {
			delete(s.entries, key)
		}
	}
	s.entries[id] = captchaEntry{answer: value, expires: s.now().Add(5 * time.Minute)}
	return nil
}

func (s *ExpiringStore) Get(id string, clear bool) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	if clear || (ok && !s.now().Before(entry.expires)) {
		delete(s.entries, id)
	}
	if !ok || !s.now().Before(entry.expires) {
		return ""
	}
	return entry.answer
}

func (s *ExpiringStore) Verify(id, answer string, clear bool) bool {
	stored := s.Get(id, true)
	return stored != "" && strings.EqualFold(stored, answer)
}

var Store base64Captcha.Store = NewExpiringStore()

type Palette struct {
	background color.RGBA
	line       color.NRGBA
	glyphs     [4]color.RGBA
}

func PaletteForTheme(theme string) Palette {
	if theme == "light" {
		return Palette{
			background: color.RGBA{R: 235, G: 238, B: 242, A: 255}, // --x-hover
			line:       color.NRGBA{R: 110, G: 115, B: 125, A: 42},
			glyphs: [4]color.RGBA{
				{R: 0, G: 98, B: 196, A: 255},  // --x-blue-fg
				{R: 54, G: 52, B: 163, A: 255}, // --x-indigo-fg
				{R: 27, G: 115, B: 48, A: 255}, // --x-green-fg
				{R: 184, G: 48, B: 0, A: 255},  // --x-orange-fg
			},
		}
	}
	return Palette{
		background: color.RGBA{R: 17, G: 24, B: 39, A: 255},
		line:       color.NRGBA{R: 150, G: 190, B: 170, A: 42},
		glyphs: [4]color.RGBA{
			{R: 248, G: 250, B: 252, A: 255},
			{R: 209, G: 250, B: 229, A: 255},
			{R: 207, G: 250, B: 254, A: 255},
			{R: 254, G: 243, B: 199, A: 255},
		},
	}
}

// 生成图形验证码
func GenerateCaptchaHandler(id string, width, height int, palette Palette) (string, error) {
	answer, err := mixedAnswer()
	if err != nil {
		return "", err
	}
	data, err := renderCaptchaPNG(answer, width, height, palette)
	if err != nil {
		return "", err
	}
	if err := Store.Set(id, answer); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
}

func renderCaptchaPNG(answer string, width, height int, palette Palette) ([]byte, error) {
	if len(answer) != 4 || width < 120 || height < 40 {
		return nil, errors.New("invalid captcha dimensions or answer")
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(palette.background), image.Point{}, draw.Src)
	drawFaintLine(canvas, 10, height/5, width-11, height/5+2, palette.line)
	drawFaintLine(canvas, 13, height*4/5, width-14, height*4/5-2, palette.line)

	face := truetype.NewFace(base64Captcha.DefaultEmbeddedFonts.LoadFontByName("fonts/wqy-microhei.ttc"), &truetype.Options{
		Size: 28, DPI: 72, Hinting: font.HintingFull,
	})
	defer face.Close()
	const margin = 12
	slotWidth := (width - 2*margin) / 4
	for i, ch := range answer {
		glyph := string(ch)
		bounds, _ := font.BoundString(face, glyph)
		centerX := margin + i*slotWidth + slotWidth/2
		centerY := height / 2
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(palette.glyphs))))
		if err != nil {
			return nil, err
		}
		drawer := font.Drawer{
			Dst:  canvas,
			Src:  image.NewUniform(palette.glyphs[index.Int64()]),
			Face: face,
			Dot: fixed.Point26_6{
				X: fixed.I(centerX) - (bounds.Min.X+bounds.Max.X)/2,
				Y: fixed.I(centerY) - (bounds.Min.Y+bounds.Max.Y)/2,
			},
		}
		drawer.DrawString(glyph)
	}
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func drawFaintLine(canvas *image.RGBA, x1, y1, x2, y2 int, line color.NRGBA) {
	for x := x1; x <= x2; x++ {
		y := y1 + (y2-y1)*(x-x1)/(x2-x1)
		draw.Draw(canvas, image.Rect(x, y, x+1, y+1), image.NewUniform(line), image.Point{}, draw.Over)
	}
}

func mixedAnswer() (string, error) {
	const digits = "23456789"
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	const all = digits + letters
	answer := make([]byte, 4)
	for i, alphabet := range []string{digits, letters, all, all} {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		answer[i] = alphabet[index.Int64()]
	}
	for i := len(answer) - 1; i > 0; i-- {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		answer[i], answer[index.Int64()] = answer[index.Int64()], answer[i]
	}
	return string(answer), nil
}

// 验证
func CaptchaVerifyHandle(id, vcode string) bool {
	return Store.Verify(id, vcode, true)
}
