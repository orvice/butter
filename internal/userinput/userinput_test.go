package userinput_test

import (
	"bytes"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/userinput"
)

// A turn keeps its parts in the order they were added: text as text, images
// as inline data with their bytes and type unchanged.
func TestTurn_KeepsPartsInOrder(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G'}
	jpeg := []byte{0xff, 0xd8, 0xff}

	var turn userinput.Turn
	for _, add := range []func() error{
		func() error { return turn.AddText("compare this") },
		func() error { return turn.AddImage("image/png", png) },
		func() error { return turn.AddText("with this") },
		func() error { return turn.AddImage("image/jpeg", jpeg) },
	} {
		if err := add(); err != nil {
			t.Fatal(err)
		}
	}

	parts := turn.Parts()
	if len(parts) != 4 {
		t.Fatalf("parts = %d, want 4", len(parts))
	}
	if parts[0].Text != "compare this" || parts[2].Text != "with this" {
		t.Errorf("text parts = %q, %q", parts[0].Text, parts[2].Text)
	}
	for i, want := range map[int]struct {
		mime string
		data []byte
	}{1: {"image/png", png}, 3: {"image/jpeg", jpeg}} {
		got := parts[i].InlineData
		if got == nil || got.MIMEType != want.mime || !bytes.Equal(got.Data, want.data) {
			t.Errorf("part %d = %+v, want %s bytes", i, parts[i], want.mime)
		}
	}
}

func TestTurn_AcceptsEveryImageType(t *testing.T) {
	for _, mime := range []string{"image/jpeg", "image/png", "image/gif", "image/webp"} {
		var turn userinput.Turn
		if err := turn.AddImage(mime, []byte{1}); err != nil {
			t.Errorf("%s: %v", mime, err)
		}
	}
}

// Each limit admits a turn exactly at it and refuses one byte, or one image,
// past it.
func TestTurn_Limits(t *testing.T) {
	type step struct {
		text  string
		mime  string
		bytes int
	}
	image := func(n int) step { return step{mime: "image/png", bytes: n} }
	images := func(count, n int) []step {
		out := make([]step, count)
		for i := range out {
			out[i] = image(n)
		}
		return out
	}

	cases := []struct {
		name    string
		steps   []step
		wantErr string
	}{
		{name: "text at the cap", steps: []step{{text: strings.Repeat("a", userinput.MaxTextBytes)}}},
		{
			name:    "text past the cap",
			steps:   []step{{text: strings.Repeat("a", userinput.MaxTextBytes+1)}},
			wantErr: "text part exceeds maximum allowed size of 1048576 bytes",
		},
		{name: "image at the cap", steps: []step{image(userinput.MaxImageBytes)}},
		{
			name:    "image past the cap",
			steps:   []step{image(userinput.MaxImageBytes + 1)},
			wantErr: "image exceeds maximum allowed size of 10485760 bytes",
		},
		{name: "as many images as allowed", steps: images(userinput.MaxImages, 1)},
		{
			name:    "one image too many",
			steps:   images(userinput.MaxImages+1, 1),
			wantErr: "too many images; maximum is 10 per request",
		},
		{name: "total at the cap", steps: images(2, userinput.MaxTotalBytes/2)},
		{
			name:    "total past the cap through text",
			steps:   append(images(2, userinput.MaxTotalBytes/2), step{text: "a"}),
			wantErr: "total parts payload exceeds maximum allowed size of 20971520 bytes",
		},
		{
			name:    "total past the cap through an image",
			steps:   append(images(2, userinput.MaxTotalBytes/2), image(1)),
			wantErr: "total parts payload exceeds maximum allowed size of 20971520 bytes",
		},
		{
			name:    "unsupported type",
			steps:   []step{{mime: "application/pdf", bytes: 1}},
			wantErr: `unsupported mime_type "application/pdf"; accepted: image/jpeg, image/png, image/gif, image/webp`,
		},
		{
			name:    "missing type",
			steps:   []step{{mime: "", bytes: 1}},
			wantErr: `unsupported mime_type ""`,
		},
		// The checks keep the RPCs' order: the count before the format, the
		// format before the size.
		{
			name:    "an extra image of an unsupported type reports the count",
			steps:   append(images(userinput.MaxImages, 1), step{mime: "application/pdf", bytes: 1}),
			wantErr: "too many images",
		},
		{
			name:    "an oversized image of an unsupported type reports the type",
			steps:   []step{{mime: "image/svg+xml", bytes: userinput.MaxImageBytes + 1}},
			wantErr: "unsupported mime_type",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var turn userinput.Turn
			var err error
			for _, s := range tc.steps {
				if s.mime != "" || s.bytes > 0 {
					err = turn.AddImage(s.mime, make([]byte, s.bytes))
				} else {
					err = turn.AddText(s.text)
				}
				if err != nil {
					break
				}
			}
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// A refused part changes nothing: it counts toward neither the total nor the
// image count, so the turn still accepts everything that fits.
func TestTurn_RefusedPartLeavesTurnUnchanged(t *testing.T) {
	var turn userinput.Turn
	if err := turn.AddImage("image/png", make([]byte, userinput.MaxImageBytes)); err != nil {
		t.Fatal(err)
	}
	if err := turn.AddImage("application/pdf", []byte{1}); err == nil {
		t.Fatal("an unsupported type was accepted")
	}
	if err := turn.AddImage("image/png", make([]byte, userinput.MaxImageBytes+1)); err == nil {
		t.Fatal("an oversized image was accepted")
	}
	if err := turn.AddText(strings.Repeat("a", userinput.MaxTextBytes+1)); err == nil {
		t.Fatal("an oversized text part was accepted")
	}
	// Exactly the total cap, had nothing refused counted.
	if err := turn.AddImage("image/png", make([]byte, userinput.MaxTotalBytes-userinput.MaxImageBytes)); err != nil {
		t.Fatalf("a refused part counted against the total: %v", err)
	}
	// Exactly the image cap, had nothing refused counted.
	for range userinput.MaxImages - 2 {
		if err := turn.AddImage("image/gif", nil); err != nil {
			t.Fatalf("a refused part counted against the images: %v", err)
		}
	}
	if got := len(turn.Parts()); got != userinput.MaxImages {
		t.Fatalf("parts = %d, want the %d accepted ones", got, userinput.MaxImages)
	}
}
