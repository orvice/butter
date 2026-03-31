package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestIsImageAttachment(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"image/png", "image/png", true},
		{"image/jpeg", "image/jpeg", true},
		{"image/gif", "image/gif", true},
		{"image/webp", "image/webp", true},
		{"IMAGE/PNG uppercase", "IMAGE/PNG", true},
		{"application/pdf", "application/pdf", false},
		{"video/mp4", "video/mp4", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &discordgo.MessageAttachment{ContentType: tt.contentType}
			if got := isImageAttachment(a); got != tt.want {
				t.Errorf("isImageAttachment(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestHasImageAttachments(t *testing.T) {
	tests := []struct {
		name        string
		attachments []*discordgo.MessageAttachment
		want        bool
	}{
		{
			name: "no attachments",
			want: false,
		},
		{
			name: "image attachment",
			attachments: []*discordgo.MessageAttachment{
				{ContentType: "image/png", Filename: "test.png"},
			},
			want: true,
		},
		{
			name: "non-image attachment",
			attachments: []*discordgo.MessageAttachment{
				{ContentType: "application/pdf", Filename: "test.pdf"},
			},
			want: false,
		},
		{
			name: "mixed attachments",
			attachments: []*discordgo.MessageAttachment{
				{ContentType: "application/pdf", Filename: "test.pdf"},
				{ContentType: "image/jpeg", Filename: "photo.jpg"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &discordgo.MessageCreate{
				Message: &discordgo.Message{
					Attachments: tt.attachments,
				},
			}
			if got := hasImageAttachments(m); got != tt.want {
				t.Errorf("hasImageAttachments() = %v, want %v", got, tt.want)
			}
		})
	}
}
