package app

import (
	"reflect"
	"testing"

	"go.orx.me/apps/butter/internal/config"
	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
)

var memoryContentStoreType = reflect.TypeOf(skillrepo.NewMemoryContentStore())

// Issue #153: without skills.s3_bucket the content store falls back to
// memory so local development needs zero infrastructure.
func TestSetupSkillContentStoreDefaultsToMemory(t *testing.T) {
	store := setupSkillContentStore(t.Context(), &config.AppConfig{})
	if got := reflect.TypeOf(store); got != memoryContentStoreType {
		t.Fatalf("expected memory content store, got %s", got)
	}
}

func TestSetupSkillContentStoreFallsBackWhenClientMissing(t *testing.T) {
	cfg := &config.AppConfig{Skills: config.SkillsConfig{
		S3Bucket:  "skill-bucket-not-registered",
		KeyPrefix: "skills",
	}}
	store := setupSkillContentStore(t.Context(), cfg)
	if got := reflect.TypeOf(store); got != memoryContentStoreType {
		t.Fatalf("expected memory fallback for unregistered s3 client, got %s", got)
	}
}
