package pgsql

import (
	"context"
	"database/sql"
	"net"

	"github.com/google/uuid"
	"github.com/leporo/sqlf"
	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/internal/repository"
)

type IPRuleRepo struct {
	pool *sql.DB
}

func (ir IPRuleRepo) Add(ctx context.Context, input model.IPRuleInput) (*model.IPRule, error) {
	guid := uuid.New()
	stmt := sqlf.InsertInto("ip_rules").
		Set("id", guid.String()).
		Set("type", input.Type.String()).
		Set("ip_net", input.IPNet.String())
	_, err := stmt.ExecAndClose(ctx, ir.pool)
	if err != nil {
		return nil, err
	}
	rules, err := ir.GetList(ctx, model.IPRuleSearch{ID: &guid})
	if err != nil {
		return nil, err
	}
	return &rules[0], nil
}

func (ir IPRuleRepo) Delete(ctx context.Context, input model.IPRuleInput) error {
	stmt := sqlf.DeleteFrom("ip_rules").
		Where("type = ?", input.Type.String()).
		Where("ip_net = ?", input.IPNet.String())
	_, err := stmt.ExecAndClose(ctx, ir.pool)
	return err
}

func (ir IPRuleRepo) GetList(ctx context.Context, search model.IPRuleSearch) ([]model.IPRule, error) {
	stmt := sqlf.From("ip_rules").
		Select(`id, type, text(ip_net), updated_at`).
		OrderBy("ip_rules.type asc")
	ir.applySearch(stmt, search)
	rules := make([]model.IPRule, 0)
	rows, err := ir.pool.QueryContext(ctx, stmt.String(), stmt.Args()...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		rule, err := ir.prepareModel(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (ir IPRuleRepo) prepareModel(row *sql.Rows) (model.IPRule, error) {
	var (
		id, typ, ipNet sql.NullString
		rule           model.IPRule
		err            error
	)
	if err := row.Scan(&id, &typ, &ipNet, &rule.UpdatedAt); err != nil {
		if err != nil {
			return rule, err
		}
	}
	if id.Valid {
		rule.ID, err = uuid.Parse(id.String)
		if err != nil {
			return rule, err
		}
	}
	if typ.Valid {
		rule.Type, err = model.ParseRuleType(typ.String)
		if err != nil {
			return rule, err
		}
	}
	if ipNet.Valid {
		_, mask, err := net.ParseCIDR(ipNet.String)
		if err != nil {
			return rule, err
		}
		rule.IPNet = *mask
	}
	return rule, nil
}

func (ir IPRuleRepo) applySearch(stmt *sqlf.Stmt, search model.IPRuleSearch) {
	if search.ID != nil {
		stmt.Where("ip_rules.id = ?", search.ID.String())
	}
	if search.Type != nil {
		stmt.Where("ip_rules.type = ?", search.Type.String())
	}
	if search.IPNet != nil {
		if search.IPNetExact {
			stmt.Where("ip_rules.ip_net = ?::inet", search.IPNet.String())
		} else {
			stmt.Where("?::inet << ip_rules.ip_net", search.IPNet.String())
		}
	}
}

func NewIPRuleRepo(pool *sql.DB) repository.IPRule {
	return &IPRuleRepo{pool: pool}
}
