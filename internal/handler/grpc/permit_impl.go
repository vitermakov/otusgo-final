package grpc

import (
	"context"
	"fmt"
	"strings"

	deps "github.com/vitermakov/otusgo-final/internal/app/deps/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/dto"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/servers/grpc/rqres"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PermitHandlerImpl расширение генерированного GRPC сервера - для публичных запросов.
type PermitHandlerImpl struct {
	pb.UnimplementedPermitServer
	services *deps.Services
	logger   logger.Logger
}

func (p PermitHandlerImpl) CheckQuery(ctx context.Context, req *pb.PermitReq) (*pb.PermitResult, error) {
	query, err := dto.PermitModel(req)
	if err != nil {
		return nil, p.handleError(fmt.Errorf("неверные данные запроса: %w", err))
	}
	res, err := p.services.PermitChecker.Check(ctx, query)
	if err != nil {
		return nil, p.handleError(fmt.Errorf("внутренняя ошибка: %w", err))
	}
	p.logCheckQuery(query, res)

	return dto.FromPermitResultModel(res), nil
}

func (p PermitHandlerImpl) ResetLogin(ctx context.Context, req *pb.RstLoginReq) (*emptypb.Empty, error) {
	query, err := dto.ResetLoginModel(req)
	if err != nil {
		return nil, p.handleError(fmt.Errorf("неверные данные сброса: %w", err))
	}
	return p.reset(ctx, query)
}

func (p PermitHandlerImpl) ResetIP(ctx context.Context, req *pb.RstIPReq) (*emptypb.Empty, error) {
	query, err := dto.ResetIPModel(req)
	if err != nil {
		return nil, p.handleError(fmt.Errorf("неверные данные сброса: %w", err))
	}
	return p.reset(ctx, query)
}

func (p PermitHandlerImpl) reset(ctx context.Context, query model.ResetQuery) (*emptypb.Empty, error) {
	_, err := p.services.PermitChecker.Reset(ctx, query)
	if err != nil {
		return nil, p.handleError(fmt.Errorf("ошибка сброса %s=%s: %w", query.Name, query.Value, err))
	}
	p.logger.Info("сброщено %s=%s", query.Name, query.Value)

	return &emptypb.Empty{}, nil
}

func (p PermitHandlerImpl) logCheckQuery(query model.PermitQuery, res model.PermitResult) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("запрос (login=%s, password=%s, ip=%s) -> ", query.Login, query.Password, query.IP))
	if res.Success {
		sb.WriteString("разрешено")
	} else {
		sb.WriteString("запрещено (" + res.Err.Error() + ")")
	}
	p.logger.Info(sb.String())
}

func (p PermitHandlerImpl) handleError(err error) error {
	p.logger.Error(err.Error())
	s := rqres.FromError(err)
	return status.Error(s.Code(), s.Message())
}
