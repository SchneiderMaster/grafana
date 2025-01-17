package sql

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/grafana/grafana/pkg/storage/unified/resource"
	"github.com/grafana/grafana/pkg/storage/unified/sql/db"
	"github.com/grafana/grafana/pkg/storage/unified/sql/dbutil"
	"github.com/grafana/grafana/pkg/storage/unified/sql/sqltemplate"
)

var (
	_ resource.BatchProcessingBackend = (*backend)(nil)
)

func (b *backend) ProcessBatch(ctx context.Context, next func() *resource.BatchRequest) error {
	return b.db.WithTx(ctx, ReadCommitted, func(ctx context.Context, tx db.Tx) error {
		var lastKey *resource.ResourceKey

		for req := next(); req != nil; req = next() {
			if req.Action == resource.BatchRequest_DELETE_COLLECTION {
				res, err := dbutil.Exec(ctx, tx, sqlResourceHistoryDelete, &sqlResourceHistoryDeleteRequest{
					SQLTemplate: sqltemplate.New(b.dialect),
					Namespace:   req.Key.Namespace,
					Group:       req.Key.Group,
					Resource:    req.Key.Resource,
				})
				if err != nil {
					return err
				}
				history_count, err := res.RowsAffected()
				if err != nil {
					return err
				}

				res, err = dbutil.Exec(ctx, tx, sqlResourceDelete, &sqlResourceRequest{
					SQLTemplate: sqltemplate.New(b.dialect),
					WriteEvent: resource.WriteEvent{
						Key: req.Key,
					},
				})
				if err != nil {
					return err
				}
				main_count, err := res.RowsAffected()
				if err != nil {
					return err
				}
				fmt.Printf("DELETE: %s (%d/%d)\n", req.Key.SearchID(), history_count, main_count)
				continue
			}

			// 4. Atomically increment resource version for this kind
			rv, err := b.resourceVersionAtomicInc(ctx, tx, req.Key)
			if err != nil {
				return fmt.Errorf("increment resource version: %w", err)
			}

			// 2. Insert into resource history
			if _, err := dbutil.Exec(ctx, tx, sqlResourceHistoryInsert, sqlResourceRequest{
				SQLTemplate: sqltemplate.New(b.dialect),
				WriteEvent: resource.WriteEvent{
					Key:        req.Key,
					Type:       resource.WatchEvent_Type(req.Action),
					PreviousRV: -1, // only used for
					Value:      req.Value,
				},
				Folder:          req.Folder,
				GUID:            uuid.NewString(),
				ResourceVersion: rv,
			}); err != nil {
				return fmt.Errorf("insert into resource history: %w", err)
			}

			lastKey = req.Key
			lastKey.Name = ""
			fmt.Printf("TODO: %s / %s / (%d)\n", req.Action, req.Key.SearchID(), rv)
		}

		// MOVE HISTORY TO MAIN
		res, err := dbutil.Exec(ctx, tx, sqlResourceInsertFromHistory, &sqlResourceInsertFromHistoryRequest{
			SQLTemplate: sqltemplate.New(b.dialect),
			Key:         lastKey, // !!!!
		})
		if err != nil {
			return err
		}
		migrate_count, err := res.RowsAffected()
		if err != nil {
			return err
		}
		fmt.Printf("SELECT: %s (%d)\n", lastKey.SearchID(), migrate_count)
		return nil
	})
}
