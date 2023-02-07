package grpc

import (
	"context"
	"fmt"

	deps "github.com/vitermakov/otusgo-final/internal/app/deps/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/dto"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/servers/grpc/rqres"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// IPRuleHandlerImpl получение разрешения на заросы, сброс бакетов.
type IPRuleHandlerImpl struct {
	pb.UnimplementedIPRuleServer
	services *deps.Services
	logger   logger.Logger
}

func (ir IPRuleHandlerImpl) AddToWhiteList(ctx context.Context, req *pb.IPNet) (*emptypb.Empty, error) {
	return ir.addToList(ctx, req, model.RuleTypeAllow)
}

func (ir IPRuleHandlerImpl) AddToBlackList(ctx context.Context, req *pb.IPNet) (*emptypb.Empty, error) {
	return ir.addToList(ctx, req, model.RuleTypeDeny)
}

func (ir IPRuleHandlerImpl) DeleteFromWhiteList(ctx context.Context, req *pb.IPNet) (*emptypb.Empty, error) {
	return ir.removeFromList(ctx, req, model.RuleTypeAllow)
}

func (ir IPRuleHandlerImpl) DeleteFromBlackList(ctx context.Context, req *pb.IPNet) (*emptypb.Empty, error) {
	return ir.removeFromList(ctx, req, model.RuleTypeDeny)
}

func (ir IPRuleHandlerImpl) addToList(
	ctx context.Context, req *pb.IPNet, ruleType model.RuleType,
) (*emptypb.Empty, error) {
	ipNet, err := dto.IPNetModel(req)
	if err != nil {
		return nil, ir.handleError(fmt.Errorf("неверно заданная сеть: %w", err))
	}
	tg := ""
	if ruleType == model.RuleTypeAllow {
		tg = "white-list"
	} else {
		tg = "black-list"
	}
	_, err = ir.services.IPRule.Add(ctx, model.IPRuleInput{IPNet: ipNet, Type: ruleType})
	if err != nil {
		return nil, ir.handleError(fmt.Errorf("ошибка добавления сети в %s: %w", tg, err))
	}
	ir.logger.Info("сеть добавлена в %s", tg)

	return &emptypb.Empty{}, nil
}

func (ir IPRuleHandlerImpl) removeFromList(
	ctx context.Context, req *pb.IPNet, ruleType model.RuleType,
) (*emptypb.Empty, error) {
	ipNet, err := dto.IPNetModel(req)
	if err != nil {
		return nil, ir.handleError(fmt.Errorf("неверно заданная сеть: %w", err))
	}
	tg := ""
	if ruleType == model.RuleTypeAllow {
		tg = "white-list"
	} else {
		tg = "black-list"
	}
	rule, err := ir.services.IPRule.GetByIPNet(ctx, ruleType, ipNet)
	if err != nil {
		ir.logger.Error(err.Error())
		s := rqres.FromError(err)
		return nil, status.Error(s.Code(), s.Message())
	}
	err = ir.services.IPRule.Delete(ctx, *rule)
	if err != nil {
		err := fmt.Errorf("ошибка удаления сети из %s: %w", tg, err)
		ir.logger.Error(err.Error())
		s := rqres.FromError(err)
		return nil, status.Error(s.Code(), s.Message())
	}
	ir.logger.Info("сеть удалена из %s", tg)
	return &emptypb.Empty{}, nil
}

func (ir IPRuleHandlerImpl) handleError(err error) error {
	ir.logger.Error(err.Error())
	s := rqres.FromError(err)
	return status.Error(s.Code(), s.Message())
}
