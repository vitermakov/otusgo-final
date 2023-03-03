package dto

import (
	"net"

	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
)

func PermitModel(req *pb.PermitReq) (model.PermitQuery, error) {
	if req == nil {
		return model.PermitQuery{}, ErrRequestEmpty
	}
	input := model.PermitQuery{
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	}
	ip := net.ParseIP(req.GetIP())
	if ip == nil {
		return model.PermitQuery{}, ErrBadIP
	}
	input.IP = ip

	return input, nil
}

func ResetLoginModel(req *pb.RstLoginReq) (model.LimitBucket, error) {
	return model.LimitBucket{Param: model.LimitParamNameLogin, Value: req.GetLogin()}, nil
}

func ResetIPModel(req *pb.RstIPReq) (model.LimitBucket, error) {
	ip := net.ParseIP(req.GetIP())
	if ip == nil {
		return model.LimitBucket{}, ErrBadIP
	}
	return model.LimitBucket{Param: model.LimitParamNameIP, Value: ip.String()}, nil
}

func FromPermitResultModel(res model.PermitResult) *pb.PermitResult {
	result := &pb.PermitResult{Success: res.Success}
	if !result.Success {
		result.Reason = res.Err.Error()
	}
	return result
}
