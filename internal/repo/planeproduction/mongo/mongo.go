package mongofactory

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo"
	planediagnostic "go.orx.me/apps/butter/internal/repo/planediagnostic/mongo"
	planeprogress "go.orx.me/apps/butter/internal/repo/planeprogress/mongo"
)

type Stores struct {
	Progress    *planeprogress.Store
	Diagnostics *planediagnostic.Store
	DB          *mongo.Database
}

func New(ctx context.Context, db *mongo.Database) (Stores, error) {
	if db == nil {
		return Stores{}, fmt.Errorf("plane durable mongo database is required")
	}
	p := planeprogress.New(db)
	d := planediagnostic.New(db)
	if e := p.EnsureIndexes(ctx); e != nil {
		return Stores{}, fmt.Errorf("plane progress indexes: %w", e)
	}
	if e := d.EnsureIndexes(ctx); e != nil {
		return Stores{}, fmt.Errorf("plane diagnostic indexes: %w", e)
	}
	return Stores{Progress: p, Diagnostics: d, DB: db}, nil
}
