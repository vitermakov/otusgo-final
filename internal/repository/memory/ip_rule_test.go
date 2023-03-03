package memory

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vitermakov/otusgo-final/internal/model"
)

func TestIPRuleMemoryRepo(t *testing.T) {
	t.Run("complex test", func(t *testing.T) {
		repo := IPRuleRepo{}
		ctx := context.Background()

		_, net1, _ := net.ParseCIDR("10.0.1.0/24")
		_, net2, _ := net.ParseCIDR("192.168.0.0/16")
		_, net3, _ := net.ParseCIDR("192.170.0.0/16")
		_, net4, _ := net.ParseCIDR("192.180.0.0/16")

		ipInNet2 := net.ParseIP("192.168.1.200")
		ipNotInNets := net.ParseIP("10.0.2.200")

		rules := []model.IPRule{
			{
				Type:  model.RuleTypeAllow,
				IPNet: *net1,
			}, {
				Type:  model.RuleTypeAllow,
				IPNet: *net2,
			}, {
				Type:  model.RuleTypeDeny,
				IPNet: *net3,
			}, {
				Type:  model.RuleTypeDeny,
				IPNet: *net4,
			},
		}
		for i, rule := range rules {
			input := model.IPRuleInput{
				Type:  rule.Type,
				IPNet: rule.IPNet,
			}
			rule, err := repo.Add(ctx, input)
			require.NoError(t, err)

			rules[i].ID = rule.ID
			rules[i].UpdatedAt = rule.UpdatedAt
		}

		actual, _ := repo.GetList(ctx, model.IPRuleSearch{})
		require.ElementsMatch(t, rules, actual)

		actual, _ = repo.GetList(ctx, model.IPRuleSearch{ID: &rules[1].ID})
		require.Equal(t, 1, len(actual))

		rt := model.RuleTypeAllow
		actual, _ = repo.GetList(ctx, model.IPRuleSearch{Type: &rt})
		require.Equal(t, 2, len(actual))
		require.ElementsMatch(t, rules[0:2], actual)

		rt = model.RuleTypeDeny
		actual, _ = repo.GetList(ctx, model.IPRuleSearch{Type: &rt})
		require.Equal(t, 2, len(actual))
		require.ElementsMatch(t, rules[2:4], actual)

		actual, _ = repo.GetList(ctx, model.IPRuleSearch{IPNet: net3, IPNetExact: true})
		require.Equal(t, 1, len(actual))
		require.Equal(t, rules[2].ID.String(), actual[0].ID.String())

		actual, _ = repo.GetList(ctx, model.IPRuleSearch{IPNet: &net.IPNet{
			IP: ipInNet2, Mask: []byte{255, 255, 255, 255},
		}})
		require.Equal(t, 1, len(actual))
		require.Equal(t, rules[1].ID.String(), actual[0].ID.String())

		actual, _ = repo.GetList(ctx, model.IPRuleSearch{IPNet: &net.IPNet{
			IP: ipNotInNets, Mask: []byte{255, 255, 255, 255},
		}, IPNetExact: true})
		require.Equal(t, 0, len(actual))

		_ = repo.Delete(ctx, model.IPRuleInput{
			Type:  rules[3].Type,
			IPNet: rules[3].IPNet,
		})
		actual, _ = repo.GetList(ctx, model.IPRuleSearch{})
		require.Equal(t, 3, len(actual))
		require.ElementsMatch(t, rules[0:3], actual)
	})
}
