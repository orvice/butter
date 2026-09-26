package memoryhook

import (
	"strings"
	"testing"
	"unicode/utf8"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/mem0"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
)

func recalled(target mem0memory.Target, text, created string) mem0memory.Recalled {
	return mem0memory.Recalled{Memory: mem0.Memory{ID: text, Memory: text, CreatedAt: created}, Target: target}
}

func TestFormatBlockSectionsAndDates(t *testing.T) {
	block := FormatBlock([]mem0memory.Recalled{
		recalled(mem0memory.TargetAgent, "answer tersely", "2026-08-20T09:00:00.5+00:00"),
		recalled(mem0memory.TargetWorkspace, "team uses pnpm", "2026-09-01T10:00:00+00:00"),
		recalled(mem0memory.TargetWorkspace, "deploys on friday", ""),
	}, MaxBlockChars)

	for _, want := range []string{
		"<memories>",
		"current conversation wins",
		"Workspace memories (shared across this workspace):\n- [2026-09-01] team uses pnpm\n- deploys on friday",
		"Agent memories (specific to you):\n- [2026-08-20] answer tersely",
		"</memories>",
	} {
		if !strings.Contains(block, want) {
			t.Fatalf("block lacks %q:\n%s", want, block)
		}
	}
	if strings.Index(block, "Workspace memories") > strings.Index(block, "Agent memories") {
		t.Fatalf("workspace section should come first:\n%s", block)
	}
}

func TestFormatBlockRespectsTheCap(t *testing.T) {
	var many []mem0memory.Recalled
	for i := 0; i < 200; i++ {
		many = append(many, recalled(mem0memory.TargetWorkspace, strings.Repeat("x", 90)+string(rune('a'+i%26)), ""))
	}
	block := FormatBlock(many, 1000)
	if n := utf8.RuneCountInString(block); n > 1000 || n == 0 {
		t.Fatalf("block runes = %d, want (0, 1000]", n)
	}
	if !strings.HasSuffix(block, "</memories>") {
		t.Fatal("capped block lost its footer")
	}
	if FormatBlock(many, 100) != "" {
		t.Fatal("a cap too small for the header should yield no block")
	}
	if FormatBlock(nil, MaxBlockChars) != "" {
		t.Fatal("no memories should yield no block")
	}
}

func TestAppendInstructionIsIdempotent(t *testing.T) {
	req := &model.LLMRequest{}
	appendInstruction(req, "BLOCK")
	if got := req.Config.SystemInstruction.Parts[0].Text; got != "BLOCK" {
		t.Fatalf("system = %q", got)
	}

	req = &model.LLMRequest{Config: &genai.GenerateContentConfig{SystemInstruction: genai.NewContentFromText("You help.", genai.RoleUser)}}
	appendInstruction(req, "BLOCK")
	appendInstruction(req, "BLOCK")
	if got := req.Config.SystemInstruction.Parts[0].Text; got != "You help.\n\nBLOCK" {
		t.Fatalf("system = %q", got)
	}
}

func testSession(t *testing.T, userTexts ...string) session.Session {
	t.Helper()
	svc := session.InMemoryService()
	created, err := svc.Create(t.Context(), &session.CreateRequest{AppName: "a", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	for i, text := range userTexts {
		ev := session.NewEvent(t.Context(), "inv")
		ev.Author = "user"
		ev.Content = genai.NewContentFromText(text, genai.RoleUser)
		if err := svc.AppendEvent(t.Context(), created.Session, ev); err != nil {
			t.Fatal(err)
		}
		reply := session.NewEvent(t.Context(), "inv")
		reply.Author = "agent"
		reply.Content = genai.NewContentFromText("reply "+string(rune('0'+i)), genai.RoleModel)
		if err := svc.AppendEvent(t.Context(), created.Session, reply); err != nil {
			t.Fatal(err)
		}
	}
	got, err := svc.Get(t.Context(), &session.GetRequest{AppName: "a", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	return got.Session
}

func TestRecallQuery(t *testing.T) {
	prior := testSession(t, "older question", "How do we deploy the web app?")
	cases := []struct {
		text  string
		prior session.Session
		want  string
	}{
		{"", prior, ""},
		{"   ", prior, ""},
		{"How should the release be tagged?", prior, "How should the release be tagged?"},
		{"ok, go on", prior, "How do we deploy the web app?\nok, go on"},
		{"ok, go on", nil, "ok, go on"},
		{"ok", testSession(t), "ok"},
	}
	for _, tc := range cases {
		if got := recallQuery(tc.text, tc.prior); got != tc.want {
			t.Errorf("recallQuery(%q) = %q, want %q", tc.text, got, tc.want)
		}
	}
}
