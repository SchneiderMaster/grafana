package dualwrite

import (
	"context"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/grafana/grafana/pkg/infra/db"
	authzextv1 "github.com/grafana/grafana/pkg/services/authz/proto/v1"
	"github.com/grafana/grafana/pkg/services/authz/zanzana"
)

func teamMembershipCollector(store db.DB) legacyTupleCollector {
	return func(ctx context.Context, orgID int64) (map[string]map[string]*openfgav1.TupleKey, error) {
		query := `
			SELECT t.uid as team_uid, u.uid as user_uid, tm.permission
			FROM team_member tm
			INNER JOIN team t ON tm.team_id = t.id
			INNER JOIN ` + store.GetDialect().Quote("user") + ` u ON tm.user_id = u.id
			WHERE t.org_id = ?
		`

		type membership struct {
			TeamUID    string `xorm:"team_uid"`
			UserUID    string `xorm:"user_uid"`
			Permission int
		}

		var memberships []membership
		err := store.WithDbSession(ctx, func(sess *db.Session) error {
			return sess.SQL(query, orgID).Find(&memberships)
		})

		if err != nil {
			return nil, err
		}

		tuples := make(map[string]map[string]*openfgav1.TupleKey)

		for _, m := range memberships {
			tuple := &openfgav1.TupleKey{
				User:   zanzana.NewTupleEntry(zanzana.TypeUser, m.UserUID, ""),
				Object: zanzana.NewTupleEntry(zanzana.TypeTeam, m.TeamUID, ""),
			}

			// Admin permission is 4 and member 0
			if m.Permission == 4 {
				tuple.Relation = zanzana.RelationTeamAdmin
			} else {
				tuple.Relation = zanzana.RelationTeamMember
			}

			if tuples[tuple.Object] == nil {
				tuples[tuple.Object] = make(map[string]*openfgav1.TupleKey)
			}

			tuples[tuple.Object][tuple.String()] = tuple
		}

		return tuples, nil
	}
}

// folderTreeCollector collects folder tree structure and writes it as relation tuples
func folderTreeCollector(store db.DB) legacyTupleCollector {
	return func(ctx context.Context, orgID int64) (map[string]map[string]*openfgav1.TupleKey, error) {
		ctx, span := tracer.Start(ctx, "accesscontrol.migrator.folderTreeCollector")
		defer span.End()

		const query = `
			SELECT uid, parent_uid, org_id FROM folder WHERE org_id = ?
		`
		type folder struct {
			FolderUID string `xorm:"uid"`
			ParentUID string `xorm:"parent_uid"`
		}

		var folders []folder
		err := store.WithDbSession(ctx, func(sess *db.Session) error {
			return sess.SQL(query, orgID).Find(&folders)
		})

		if err != nil {
			return nil, err
		}

		tuples := make(map[string]map[string]*openfgav1.TupleKey)

		for _, f := range folders {
			var tuple *openfgav1.TupleKey
			if f.ParentUID == "" {
				continue
			}

			tuple = &openfgav1.TupleKey{
				Object:   zanzana.NewTupleEntry(zanzana.TypeFolder, f.FolderUID, ""),
				Relation: zanzana.RelationParent,
				User:     zanzana.NewTupleEntry(zanzana.TypeFolder, f.ParentUID, ""),
			}

			if tuples[tuple.Object] == nil {
				tuples[tuple.Object] = make(map[string]*openfgav1.TupleKey)
			}

			tuples[tuple.Object][tuple.String()] = tuple
		}

		return tuples, nil
	}
}

// managedPermissionsCollector collects managed permissions.
// It will only store actions that are supported by our schema. Managed permissions can
// be directly mapped to user/team/role without having to write an intermediate role.
func managedPermissionsCollector(store db.DB, kind string) legacyTupleCollector {
	return func(ctx context.Context, orgID int64) (map[string]map[string]*openfgav1.TupleKey, error) {
		query := `
			SELECT u.uid as user_uid, t.uid as team_uid, p.action, p.kind, p.identifier, r.org_id
			FROM permission p
			INNER JOIN role r ON p.role_id = r.id
			LEFT JOIN user_role ur ON r.id = ur.role_id
			LEFT JOIN ` + store.GetDialect().Quote("user") + ` u ON u.id = ur.user_id
			LEFT JOIN team_role tr ON r.id = tr.role_id
			LEFT JOIN team t ON tr.team_id = t.id
			LEFT JOIN builtin_role br ON r.id  = br.role_id
			WHERE r.name LIKE 'managed:%'
			AND r.org_id = ?
			AND p.kind = ?
		`
		type Permission struct {
			RoleName   string `xorm:"role_name"`
			Action     string `xorm:"action"`
			Kind       string
			Identifier string
			UserUID    string `xorm:"user_uid"`
			TeamUID    string `xorm:"team_uid"`
		}

		var permissions []Permission
		err := store.WithDbSession(ctx, func(sess *db.Session) error {
			return sess.SQL(query, orgID, kind).Find(&permissions)
		})

		if err != nil {
			return nil, err
		}

		tuples := make(map[string]map[string]*openfgav1.TupleKey)

		for _, p := range permissions {
			var subject string
			if len(p.UserUID) > 0 {
				subject = zanzana.NewTupleEntry(zanzana.TypeUser, p.UserUID, "")
			} else if len(p.TeamUID) > 0 {
				subject = zanzana.NewTupleEntry(zanzana.TypeTeam, p.TeamUID, "member")
			} else {
				// FIXME(kalleep): Unsuported role binding (org role). We need to have basic roles in place
				continue
			}

			tuple, ok := zanzana.TranslateToResourceTuple(subject, p.Action, p.Kind, p.Identifier)
			if !ok {
				continue
			}

			if tuples[tuple.Object] == nil {
				tuples[tuple.Object] = make(map[string]*openfgav1.TupleKey)
			}

			// For resource actions on folders we need to merge the tuples into one with combined
			// group_resources.
			if zanzana.IsFolderResourceTuple(tuple) {
				key := tupleStringWithoutCondition(tuple)
				if t, ok := tuples[tuple.Object][key]; ok {
					zanzana.MergeFolderResourceTuples(t, tuple)
				} else {
					tuples[tuple.Object][key] = tuple
				}

				continue
			}

			tuples[tuple.Object][tuple.String()] = tuple
		}

		return tuples, nil
	}
}

func tupleStringWithoutCondition(tuple *openfgav1.TupleKey) string {
	c := tuple.Condition
	tuple.Condition = nil
	s := tuple.String()
	tuple.Condition = c
	return s
}

// basicRolePermissionsCollector collects permissions for basic roles
func basicRolePermissionsCollector(store db.DB) legacyTupleCollector {
	return func(ctx context.Context, _ int64) (map[string]map[string]*openfgav1.TupleKey, error) {
		const query = `
			SELECT r.uid as role_uid, p.action, p.kind, p.identifier
			FROM permission p
			INNER JOIN role r ON p.role_id = r.id
			LEFT JOIN builtin_role br ON r.id  = br.role_id
			WHERE r.name LIKE 'basic:%'
		`
		type Permission struct {
			Action     string `xorm:"action"`
			Kind       string
			Identifier string
			RoleUID    string `xorm:"role_uid"`
		}

		var permissions []Permission
		err := store.WithDbSession(ctx, func(sess *db.Session) error {
			return sess.SQL(query).Find(&permissions)
		})
		if err != nil {
			return nil, err
		}

		tuples := make(map[string]map[string]*openfgav1.TupleKey)

		for _, p := range permissions {
			subject := zanzana.NewTupleEntry(zanzana.TypeRole, p.RoleUID, "assignee")

			tuple, ok := zanzana.TranslateToResourceTuple(subject, p.Action, p.Kind, p.Identifier)
			if !ok {
				continue
			}

			if tuples[tuple.Object] == nil {
				tuples[tuple.Object] = make(map[string]*openfgav1.TupleKey)
			}

			tuples[tuple.Object][tuple.String()] = tuple
		}

		return tuples, nil
	}
}

// basicRoleBindingsCollects collects role bindings for basic roles
func basicRoleBindingsCollector(store db.DB) legacyTupleCollector {
	return func(ctx context.Context, orgID int64) (map[string]map[string]*openfgav1.TupleKey, error) {
		query := `
			SELECT ou.org_id, u.uid as user_uid, ou.role as org_role
			FROM org_user ou
			LEFT JOIN ` + store.GetDialect().Quote("user") + ` u ON u.id = ou.user_id
			WHERE ou.org_id = ?
			AND NOT u.is_service_account
		`
		// FIXME: handle service admin role
		type Binding struct {
			UserUID string `xorm:"user_uid"`
			OrgRole string `xorm:"org_role"`
		}

		var bindings []Binding
		err := store.WithDbSession(ctx, func(sess *db.Session) error {
			return sess.SQL(query, orgID).Find(&bindings)
		})

		if err != nil {
			return nil, err
		}

		tuples := make(map[string]map[string]*openfgav1.TupleKey)

		for _, b := range bindings {
			subject := zanzana.NewTupleEntry(zanzana.TypeUser, b.UserUID, "")

			tuple := &openfgav1.TupleKey{
				User:     subject,
				Relation: zanzana.RelationAssignee,
				Object:   zanzana.NewTupleEntry(zanzana.TypeRole, zanzana.TranslateBasicRole(b.OrgRole), ""),
			}

			if tuples[tuple.Object] == nil {
				tuples[tuple.Object] = make(map[string]*openfgav1.TupleKey)
			}

			tuples[tuple.Object][tuple.String()] = tuple
		}

		return tuples, nil
	}
}

func zanzanaCollector(relations []string) zanzanaTupleCollector {
	return func(ctx context.Context, client zanzana.Client, object string, namespace string) (map[string]*openfgav1.TupleKey, error) {
		// list will use continuation token to collect all tuples for object and relation
		list := func(relation string) ([]*openfgav1.Tuple, error) {
			first, err := client.Read(ctx, &authzextv1.ReadRequest{
				Namespace: namespace,
				TupleKey: &authzextv1.ReadRequestTupleKey{
					Object:   object,
					Relation: relation,
				},
			})

			if err != nil {
				return nil, err
			}

			c := first.ContinuationToken

			for c != "" {
				res, err := client.Read(ctx, &authzextv1.ReadRequest{
					Namespace: namespace,
					TupleKey: &authzextv1.ReadRequestTupleKey{
						Object:   object,
						Relation: relation,
					},
				})
				if err != nil {
					return nil, err
				}

				c = res.ContinuationToken
				first.Tuples = append(first.Tuples, res.Tuples...)
			}

			return zanzana.ToOpenFGATuples(first.Tuples), nil
		}

		out := make(map[string]*openfgav1.TupleKey)
		for _, r := range relations {
			tuples, err := list(r)
			if err != nil {
				return nil, err
			}
			for _, t := range tuples {
				if zanzana.IsFolderResourceTuple(t.Key) {
					out[tupleStringWithoutCondition(t.Key)] = t.Key
				} else {
					out[t.Key.String()] = t.Key
				}
			}
		}

		return out, nil
	}
}
