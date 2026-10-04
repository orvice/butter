// Package userinput holds the rules for what one user turn may carry into a
// run: text and inline images, within fixed limits. Every entry point that
// accepts them builds the turn through Turn, so the AgentService and
// SessionService RPCs and the AG-UI endpoint accept exactly the same input.
//
// Errors are plain errors. Each caller maps them onto its own protocol: the
// RPCs answer invalid_argument, the AG-UI endpoint answers 400.
package userinput

import (
	"fmt"
	"slices"
	"strings"

	"google.golang.org/genai"
)

const (
	// MaxTextBytes caps one text part. The single-message request fields
	// that predate parts carry the same cap, so no input path is unbounded.
	MaxTextBytes = 1 << 20 // 1 MiB

	// MaxImageBytes caps one inline image. It stays under typical
	// model-provider inline-data ceilings while protecting server memory
	// under concurrent load.
	MaxImageBytes = 10 << 20 // 10 MiB

	// MaxImages caps how many inline images one turn may carry.
	MaxImages = 10

	// MaxTotalBytes caps one turn's combined payload: its text plus its
	// image bytes.
	MaxTotalBytes = 20 << 20 // 20 MiB
)

// imageTypes are the accepted inline image formats, in the order errors list
// them.
var imageTypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

var errTotalTooLarge = fmt.Errorf("total parts payload exceeds maximum allowed size of %d bytes", MaxTotalBytes)

// Turn collects one user turn's parts in order and checks each against the
// limits as it is added. A part that breaks a limit is refused with an error
// and leaves the turn unchanged. The zero value is an empty turn.
type Turn struct {
	parts  []*genai.Part
	total  int
	images int
}

// AddText adds a text part.
func (t *Turn) AddText(text string) error {
	if len(text) > MaxTextBytes {
		return fmt.Errorf("text part exceeds maximum allowed size of %d bytes", MaxTextBytes)
	}
	total := t.total + len(text)
	if total > MaxTotalBytes {
		return errTotalTooLarge
	}
	t.total = total
	t.parts = append(t.parts, genai.NewPartFromText(text))
	return nil
}

// AddImage adds an inline image part. The checks run in a fixed order: the
// image count, then the format, then the image's size, then the turn's total.
func (t *Turn) AddImage(mimeType string, data []byte) error {
	if t.images >= MaxImages {
		return fmt.Errorf("too many images; maximum is %d per request", MaxImages)
	}
	if !slices.Contains(imageTypes, mimeType) {
		return fmt.Errorf("unsupported mime_type %q; accepted: %s", mimeType, strings.Join(imageTypes, ", "))
	}
	if len(data) > MaxImageBytes {
		return fmt.Errorf("image exceeds maximum allowed size of %d bytes", MaxImageBytes)
	}
	total := t.total + len(data)
	if total > MaxTotalBytes {
		return errTotalTooLarge
	}
	t.images++
	t.total = total
	t.parts = append(t.parts, genai.NewPartFromBytes(data, mimeType))
	return nil
}

// Parts returns the parts added so far, in the order they were added.
func (t *Turn) Parts() []*genai.Part {
	return t.parts
}
