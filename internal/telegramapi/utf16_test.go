package telegramapi

import "testing"

func TestSliceUTF16UsesTelegramEntityCoordinates(t *testing.T) {
	text := "🙂 @opsbot hello"
	got, ok := SliceUTF16(text, 3, 7)
	if !ok || got != "@opsbot" {
		t.Fatalf("SliceUTF16 = %q, %v", got, ok)
	}
	if _, ok := SliceUTF16(text, 1, 1); ok {
		t.Fatal("accepted an offset inside an emoji surrogate pair")
	}
}

func TestCommandUsesUTF16EntityLength(t *testing.T) {
	msg := &Message_{
		Text:     "/deploy🙂 staging",
		Entities: []MessageEntity{{Type: "bot_command", Offset: 0, Length: 9}},
	}
	command, args, ok := Command(msg)
	if !ok || command != "deploy🙂" || args != "staging" {
		t.Fatalf("Command = %q, %q, %v", command, args, ok)
	}
}
