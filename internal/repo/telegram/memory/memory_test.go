package memory

import (
	"errors"
	"testing"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func newChannel(id, key, botID string) *agentsv1.TelegramChannel {
	return &agentsv1.TelegramChannel{Id: id, Key: key, Name: key, BotId: botID}
}

func newDestination(id, key, channelID, chatID, threadID string) *agentsv1.TelegramDestination {
	return &agentsv1.TelegramDestination{
		Id: id, Key: key, ChannelId: channelID, ChatId: chatID, MessageThreadId: threadID,
	}
}

func seedChannel(t *testing.T, store *Store, workspaceID, id, key, botID string) *agentsv1.TelegramChannel {
	t.Helper()
	created, err := store.CreateChannel(t.Context(), workspaceID, newChannel(id, key, botID),
		telegramrepo.Credential{Ciphertext: "cipher-" + id, KeyID: "key-1"})
	if err != nil {
		t.Fatalf("CreateChannel(%s): %v", id, err)
	}
	return created
}

func TestCreateChannelStampsRevisionAndCredentialState(t *testing.T) {
	store := New()
	created := seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	if created.GetRevision() != 1 {
		t.Errorf("revision = %d, want 1", created.GetRevision())
	}
	if created.GetCredentialState() == agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_MISSING {
		t.Error("expected a stored credential to be reported as present")
	}
	if created.GetCredentialUpdatedAt() == nil {
		t.Error("expected credential_updated_at to be stamped")
	}
	if created.GetWorkspaceId() != "ws" {
		t.Errorf("workspace_id = %q", created.GetWorkspaceId())
	}
}

// A Channel read must never carry credential material, and a Channel with no
// stored token must report MISSING regardless of what the spec said.
func TestChannelReadsNeverExposeCredentials(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	if err := store.SetChannelCredential(t.Context(), "ws", "ch-1", telegramrepo.Credential{}); err != nil {
		t.Fatalf("SetChannelCredential(clear): %v", err)
	}
	got, err := store.GetChannel(t.Context(), "ws", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if got.GetCredentialState() != agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_MISSING {
		t.Errorf("credential_state = %v, want MISSING", got.GetCredentialState())
	}
	if got.GetCredentialUpdatedAt() != nil {
		t.Error("expected credential_updated_at to be cleared")
	}
	if _, err := store.GetChannelCredential(t.Context(), "ws", "ch-1"); !errors.Is(err, telegramrepo.ErrNoCredential) {
		t.Errorf("err = %v, want ErrNoCredential", err)
	}
}

func TestChannelKeyIsUniquePerWorkspace(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	_, err := store.CreateChannel(t.Context(), "ws", newChannel("ch-2", "main-bot", "43"), telegramrepo.Credential{})
	if !errors.Is(err, telegramrepo.ErrKeyExists) {
		t.Fatalf("err = %v, want ErrKeyExists", err)
	}
	// The same key in another workspace is fine — keys are workspace-scoped.
	if _, err := store.CreateChannel(t.Context(), "other", newChannel("ch-3", "main-bot", "44"), telegramrepo.Credential{}); err != nil {
		t.Fatalf("CreateChannel in other workspace: %v", err)
	}
}

// Bot IDs are global, not workspace-scoped: two Channels on one Bot would
// both consume its updates.
func TestBotIDIsUniqueAcrossWorkspaces(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	_, err := store.CreateChannel(t.Context(), "other", newChannel("ch-2", "other-bot", "42"), telegramrepo.Credential{})
	if !errors.Is(err, telegramrepo.ErrBotExists) {
		t.Fatalf("err = %v, want ErrBotExists", err)
	}
}

func TestUpdateChannelRequiresTheExpectedRevision(t *testing.T) {
	store := New()
	created := seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	created.Name = "Renamed"
	updated, err := store.UpdateChannel(t.Context(), "ws", created, 1)
	if err != nil {
		t.Fatalf("UpdateChannel: %v", err)
	}
	if updated.GetRevision() != 2 {
		t.Errorf("revision = %d, want 2", updated.GetRevision())
	}

	// Replaying the stale revision must not overwrite the newer state.
	stale := newChannel("ch-1", "main-bot", "42")
	stale.Name = "Clobbered"
	if _, err := store.UpdateChannel(t.Context(), "ws", stale, 1); !errors.Is(err, telegramrepo.ErrRevisionConflict) {
		t.Fatalf("err = %v, want ErrRevisionConflict", err)
	}
	current, err := store.GetChannel(t.Context(), "ws", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if current.GetName() != "Renamed" {
		t.Errorf("name = %q, want the value from the winning write", current.GetName())
	}
}

func TestUpdateChannelPreservesImmutableFields(t *testing.T) {
	store := New()
	created := seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	created.Key = "hijacked"
	created.BotId = "999"
	updated, err := store.UpdateChannel(t.Context(), "ws", created, 1)
	if err != nil {
		t.Fatalf("UpdateChannel: %v", err)
	}
	if updated.GetKey() != "main-bot" {
		t.Errorf("key = %q, want the stored key", updated.GetKey())
	}
	if updated.GetBotId() != "42" {
		t.Errorf("bot_id = %q, want the stored bot id", updated.GetBotId())
	}
}

func TestRotateChannelCredentialIsAtomicWithTheRevision(t *testing.T) {
	store := New()
	created := seedChannel(t, store, "ws", "ch-1", "main-bot", "42")
	before, err := store.GetChannelCredential(t.Context(), "ws", "ch-1")
	if err != nil {
		t.Fatalf("GetChannelCredential: %v", err)
	}

	created.BotUsername = "renamed"
	if _, err := store.RotateChannelCredential(t.Context(), "ws", created,
		telegramrepo.Credential{Ciphertext: "new-cipher", KeyID: "key-2"}, 0); !errors.Is(err, telegramrepo.ErrRevisionConflict) {
		t.Fatalf("err = %v, want ErrRevisionConflict", err)
	}
	after, err := store.GetChannelCredential(t.Context(), "ws", "ch-1")
	if err != nil {
		t.Fatalf("GetChannelCredential after conflict: %v", err)
	}
	if after != before {
		t.Fatalf("credential changed after failed rotation: before=%+v after=%+v", before, after)
	}
	current, err := store.GetChannel(t.Context(), "ws", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if current.GetBotUsername() == "renamed" {
		t.Fatal("channel metadata changed after failed rotation")
	}
}

func TestDeleteChannelIsBlockedWhileADestinationReferencesIt(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")
	if _, err := store.CreateDestination(t.Context(), "ws", newDestination("d-1", "ops", "ch-1", "-100", "")); err != nil {
		t.Fatalf("CreateDestination: %v", err)
	}

	if err := store.DeleteChannel(t.Context(), "ws", "ch-1"); !errors.Is(err, telegramrepo.ErrChannelInUse) {
		t.Fatalf("err = %v, want ErrChannelInUse", err)
	}
	if err := store.DeleteDestination(t.Context(), "ws", "d-1"); err != nil {
		t.Fatalf("DeleteDestination: %v", err)
	}
	if err := store.DeleteChannel(t.Context(), "ws", "ch-1"); err != nil {
		t.Fatalf("DeleteChannel after removing the reference: %v", err)
	}
}

// One inbound update must never match two Destinations, so the exact
// (channel, chat, thread) address is unique.
func TestDestinationAddressIsUnique(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")
	if _, err := store.CreateDestination(t.Context(), "ws", newDestination("d-1", "ops", "ch-1", "-100", "7")); err != nil {
		t.Fatalf("CreateDestination: %v", err)
	}

	_, err := store.CreateDestination(t.Context(), "ws", newDestination("d-2", "ops-dup", "ch-1", "-100", "7"))
	if !errors.Is(err, telegramrepo.ErrAddressExists) {
		t.Fatalf("err = %v, want ErrAddressExists", err)
	}
}

// An absent thread ID is a distinct address from any real thread, not a
// wildcard over the group.
func TestNonTopicAddressIsDistinctFromTopicAddresses(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")
	if _, err := store.CreateDestination(t.Context(), "ws", newDestination("d-1", "general", "ch-1", "-100", "")); err != nil {
		t.Fatalf("CreateDestination(general): %v", err)
	}
	if _, err := store.CreateDestination(t.Context(), "ws", newDestination("d-2", "topic", "ch-1", "-100", "7")); err != nil {
		t.Fatalf("CreateDestination(topic): %v", err)
	}

	general, err := store.FindDestinationByAddress(t.Context(), "ch-1", "-100", "")
	if err != nil {
		t.Fatalf("FindDestinationByAddress(general): %v", err)
	}
	if general.GetId() != "d-1" {
		t.Errorf("general resolved to %q", general.GetId())
	}
	topic, err := store.FindDestinationByAddress(t.Context(), "ch-1", "-100", "7")
	if err != nil {
		t.Fatalf("FindDestinationByAddress(topic): %v", err)
	}
	if topic.GetId() != "d-2" {
		t.Errorf("topic resolved to %q", topic.GetId())
	}
	if _, err := store.FindDestinationByAddress(t.Context(), "ch-1", "-100", "9"); !errors.Is(err, telegramrepo.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound for an unconfigured topic", err)
	}
}

func TestUpdateDestinationPreservesTheAddress(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")
	created, err := store.CreateDestination(t.Context(), "ws", newDestination("d-1", "ops", "ch-1", "-100", "7"))
	if err != nil {
		t.Fatalf("CreateDestination: %v", err)
	}

	created.ChatId = "-200"
	created.MessageThreadId = "9"
	created.Name = "Renamed"
	updated, err := store.UpdateDestination(t.Context(), "ws", created, 1)
	if err != nil {
		t.Fatalf("UpdateDestination: %v", err)
	}
	if updated.GetChatId() != "-100" || updated.GetMessageThreadId() != "7" {
		t.Errorf("address changed to chat %q thread %q", updated.GetChatId(), updated.GetMessageThreadId())
	}
	if updated.GetName() != "Renamed" {
		t.Errorf("name = %q, want the updated value", updated.GetName())
	}
}

func TestWorkspaceIsolation(t *testing.T) {
	store := New()
	seedChannel(t, store, "ws", "ch-1", "main-bot", "42")

	if _, err := store.GetChannel(t.Context(), "other", "ch-1"); !errors.Is(err, telegramrepo.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound across workspaces", err)
	}
	// The receive path has no workspace context, so an unscoped lookup must
	// still resolve.
	if _, err := store.FindChannel(t.Context(), "ch-1"); err != nil {
		t.Fatalf("FindChannel: %v", err)
	}
}
