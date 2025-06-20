package font

import (
	"log"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type Font struct {
	Atlas      *sdl.Surface
	GridWidth  int
	CharSize   [2]int // Width and height of each character cell (excluding 1px padding)
	CharSet    string // String containing all supported characters in order matching atlas
	CharWidths []int  // Width of each character (indices match CharSet)
	NewlinePad int    // Extra vertical padding between lines
	LetterPad  int    // Extra horizontal padding between characters
}

// gets the glyph rect from the atlas
func (font *Font) loadGlyph(char rune) sdl.Rect {
	index := strings.IndexRune(font.CharSet, char)
	if index < 0 {
		log.Printf("Character %q not found in font charset", char)
		return sdl.Rect{X: 0, Y: 0, W: 0, H: 0}
	}

	gridX := int32((index % font.GridWidth) * (font.CharSize[0] + 1)) // +1 for padding
	gridY := int32((index / font.GridWidth) * (font.CharSize[1] + 1)) // +1 for padding
	width := int32(font.CharWidths[index])
	height := int32(font.CharSize[1])

	return sdl.Rect{X: gridX, Y: gridY, W: width, H: height}
}

// renderString draws text to the screen with specified position and color
func (font *Font) RenderString(text string, r, g, b float64) *sdl.Surface { // 0-1 rgb color
	cursorX := 0
	cursorY := 0

	surface, err := sdl.CreateRGBSurface(
		0,
		int32(font.getStringLen(text)),
		int32(font.CharSize[1]),
		32,
		0x00FF0000,
		0x0000FF00,
		0x000000FF,
		0xFF000000,
	)
	if err != nil {
		panic(err)
	}

	// set modulation
	font.Atlas.SetColorMod(uint8(r*255), uint8(g*255), uint8(b*255))

	for _, char := range text {
		if char == '\n' {
			// Handle newlines
			cursorX = 0
			cursorY += font.CharSize[1] + font.NewlinePad
			continue
		}

		srcRect := font.loadGlyph(char)
		if srcRect.W == 0 || srcRect.H == 0 {
			// skip unknown chars
			continue
		}

		dstRect := sdl.Rect{X: int32(cursorX), Y: int32(cursorY), W: srcRect.W, H: srcRect.H}

		// blit
		font.Atlas.Blit(
			&srcRect,
			surface,
			&dstRect,
		)

		// Advance cursor
		charWidth := font.CharWidths[strings.IndexRune(font.CharSet, char)]
		cursorX += charWidth + font.LetterPad
	}
	return surface
}

func (font Font) getStringLen(text string) (ln int) {
	for i := range text {
		// Advance cursor
		charWidth := font.CharWidths[strings.IndexRune(font.CharSet, rune(text[i]))]
		ln += charWidth + font.LetterPad
	}
	return
}

// newFont creates a new Font from an atlas image file
func NewFont(atlasPath string, gridWidth int, charSet string, charWidths []int) (font Font) {
	// Load font atlas image
	surface, err := img.Load(atlasPath)
	if err != nil {
		panic(err)
	}

	font = Font{
		Atlas:      surface,
		GridWidth:  gridWidth,
		CharSize:   [2]int{5, 11}, // Most chars are 5x7, some extend below baseline to 11px
		CharSet:    charSet,
		CharWidths: charWidths,
		LetterPad:  1,
		NewlinePad: 5,
	}

	return
}

// Default font using the Isometrica typeface
func MakeDefaultFont() Font {
	return NewFont(
		"font/assets/font_atlas.png",
		10, // Characters per row in atlas
		" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]\\^_`abcdefghijklmnopqrstuvwxyz{}|~", // add support for ⟨⟩⟪⟫☺
		// Character widths (matching order of CharSet above):
		[]int{
			3,                            // Space
			1,                            // !
			3,                            // "
			5,                            // #
			5,                            // $
			5,                            // %
			5,                            // &
			1,                            // '
			2,                            // (
			2,                            // )
			3,                            // *
			3,                            // +
			1,                            // ,
			3,                            // -
			1,                            // .
			5,                            // /
			5, 3, 5, 5, 5, 5, 5, 5, 5, 5, // 0-9
			1,                                                                            // :
			1,                                                                            // ;
			3,                                                                            // <
			3,                                                                            // =
			3,                                                                            // >
			4,                                                                            // ?
			5,                                                                            // @
			5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, // A-Z
			2,                                                                            // [
			2,                                                                            // \
			5,                                                                            // ]
			3,                                                                            // ^
			3,                                                                            // _
			2,                                                                            // `
			5, 5, 4, 5, 5, 4, 5, 5, 1, 4, 4, 3, 5, 4, 4, 5, 5, 4, 4, 4, 5, 3, 5, 3, 4, 4, // a-z
			3, // {
			3, // |
			1, // }
			4, // ~
		},
	)
}
