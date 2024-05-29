package notifier

import (
	"context"
	"math/rand"
	"testing"

	"github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	alertingmodels "github.com/grafana/alerting/models"
	ngfakes "github.com/grafana/grafana/pkg/services/ngalert/tests/fakes"

	ac "github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/auth/identity"
	"github.com/grafana/grafana/pkg/services/ngalert/accesscontrol"
	"github.com/grafana/grafana/pkg/services/ngalert/accesscontrol/fakes"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/util"
)

func TestWithAccessControlMetadata(t *testing.T) {
	user := ac.BackgroundUser("test", 1, org.RoleNone, nil)
	silencesWithMetadata := []*models.SilenceWithMetadata{
		{Silence: util.Pointer(models.SilenceGen()())},
		{Silence: util.Pointer(models.SilenceGen()())},
		{Silence: util.Pointer(models.SilenceGen()())},
	}
	randPerm := func() models.SilencePermissionSet {
		return models.SilencePermissionSet{
			models.SilencePermissionRead:   rand.Intn(2) == 1,
			models.SilencePermissionWrite:  rand.Intn(2) == 1,
			models.SilencePermissionCreate: rand.Intn(2) == 1,
		}
	}
	t.Run("Attach permissions to silences", func(t *testing.T) {
		authz := fakes.FakeSilenceService{}
		response := map[*models.Silence]models.SilencePermissionSet{
			silencesWithMetadata[0].Silence: randPerm(),
			silencesWithMetadata[1].Silence: randPerm(),
			silencesWithMetadata[2].Silence: randPerm(),
		}
		authz.SilenceAccessFunc = func(ctx context.Context, user identity.Requester, silences []*models.Silence) (map[*models.Silence]models.SilencePermissionSet, error) {
			return response, nil
		}
		svc := SilenceService{
			authz: &authz,
		}

		require.NoError(t, svc.WithAccessControlMetadata(context.Background(), user, silencesWithMetadata...))
		for _, silence := range silencesWithMetadata {
			assert.Equal(t, response[silence.Silence], *silence.Metadata.Permissions)
		}
	})
}

func TestWithRuleMetadata(t *testing.T) {
	user := ac.BackgroundUser("test", 1, org.RoleNone, nil)
	t.Run("Attach rule metadata to silences", func(t *testing.T) {
		ruleAuthz := fakes.FakeRuleService{}
		ruleAuthz.HasAccessInFolderFunc = func(ctx context.Context, user identity.Requester, silence accesscontrol.Namespaced) (bool, error) {
			return true, nil
		}

		rules := []*models.AlertRule{
			{UID: "rule1", NamespaceUID: "folder1"},
			{UID: "rule2", NamespaceUID: "folder2"},
			{UID: "rule3", NamespaceUID: "folder3"},
		}
		ruleStore := ngfakes.NewRuleStore(t)
		ruleStore.Rules[1] = rules
		svc := SilenceService{
			ruleAuthz: &ruleAuthz,
			ruleStore: ruleStore,
		}

		silencesWithMetadata := []*models.SilenceWithMetadata{
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule1", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule2", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule3", labels.MatchEqual))())},
		}

		require.NoError(t, svc.WithRuleMetadata(context.Background(), user, silencesWithMetadata...))
		for i, silence := range silencesWithMetadata {
			metadata := &models.SilenceRuleMetadata{
				RuleUID:   rules[i].UID,
				RuleTitle: rules[i].Title,
				FolderUID: rules[i].NamespaceUID,
			}
			assert.Equal(t, silence.Metadata, models.SilenceMetadata{RuleMetadata: metadata})
		}
	})
	t.Run("Don't attach rule metadata if no access or global", func(t *testing.T) {
		ruleAuthz := fakes.FakeRuleService{}
		ruleAuthz.HasAccessInFolderFunc = func(ctx context.Context, user identity.Requester, silence accesscontrol.Namespaced) (bool, error) {
			return silence.GetNamespaceUID() == "folder1", nil
		}

		rules := []*models.AlertRule{
			{UID: "rule1", NamespaceUID: "folder1"},
			{UID: "rule2", NamespaceUID: "folder2"},
			{UID: "rule3", NamespaceUID: "folder3"},
		}
		ruleStore := ngfakes.NewRuleStore(t)
		ruleStore.Rules[1] = rules
		svc := SilenceService{
			ruleAuthz: &ruleAuthz,
			ruleStore: ruleStore,
		}

		silencesWithMetadata := []*models.SilenceWithMetadata{
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule1", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule2", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule3", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen()())},
		}

		require.NoError(t, svc.WithRuleMetadata(context.Background(), user, silencesWithMetadata...))
		for i, silence := range silencesWithMetadata {
			var metadata *models.SilenceRuleMetadata
			if silence.GetRuleUID() != nil && *silence.GetRuleUID() == "rule1" {
				metadata = &models.SilenceRuleMetadata{
					RuleUID:   rules[i].UID,
					RuleTitle: rules[i].Title,
					FolderUID: rules[i].NamespaceUID,
				}
			}
			assert.Equal(t, silence.Metadata, models.SilenceMetadata{RuleMetadata: metadata})
		}
	})
	t.Run("Don't check same namespace access more than once", func(t *testing.T) {
		ruleAuthz := fakes.FakeRuleService{}
		ruleAuthz.HasAccessInFolderFunc = func(ctx context.Context, user identity.Requester, silence accesscontrol.Namespaced) (bool, error) {
			return true, nil
		}

		rules := []*models.AlertRule{
			{UID: "rule1", NamespaceUID: "folder1"},
			{UID: "rule2", NamespaceUID: "folder1"},
			{UID: "rule3", NamespaceUID: "folder1"},
		}
		ruleStore := ngfakes.NewRuleStore(t)
		ruleStore.Rules[1] = rules
		svc := SilenceService{
			ruleAuthz: &ruleAuthz,
			ruleStore: ruleStore,
		}

		silencesWithMetadata := []*models.SilenceWithMetadata{
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule1", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule2", labels.MatchEqual))())},
			{Silence: util.Pointer(models.SilenceGen(models.SilenceMuts.WithMatcher(alertingmodels.RuleUIDLabel, "rule3", labels.MatchEqual))())},
		}

		require.NoError(t, svc.WithRuleMetadata(context.Background(), user, silencesWithMetadata...))
		assert.Lenf(t, ruleAuthz.Calls, 1, "HasAccessInFolder should be called only once per namespace")
		assert.Equal(t, "HasAccessInFolder", ruleAuthz.Calls[0].MethodName)
		assert.Equal(t, "folder1", ruleAuthz.Calls[0].Arguments[2].(accesscontrol.Namespaced).GetNamespaceUID())
	})
}
