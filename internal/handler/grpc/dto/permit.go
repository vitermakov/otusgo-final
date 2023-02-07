package dto

import (
	"errors"
	"net"

	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
)

func PermitModel(req *pb.PermitReq) (model.PermitQuery, error) {
	if req == nil {
		return model.PermitQuery{}, errors.New("empty query")
	}
	input := model.PermitQuery{
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	}
	ip := net.ParseIP(req.GetIP())
	if ip == nil {
		return model.PermitQuery{}, errors.New("ip address is not well-formed")
	}
	input.IP = ip

	return input, nil
}

func ResetLoginModel(req *pb.RstLoginReq) (model.ResetQuery, error) {
	return model.ResetQuery{Name: "login", Value: req.GetLogin()}, nil
}

func ResetIPModel(req *pb.RstIPReq) (model.ResetQuery, error) {
	ip := net.ParseIP(req.GetIP())
	if ip == nil {
		return model.ResetQuery{}, errors.New("ip address is not well-formed")
	}
	return model.ResetQuery{Name: "ip", Value: ip.String()}, nil
}

func FromPermitResultModel(res model.PermitResult) *pb.PermitResult {
	result := &pb.PermitResult{Success: res.Success}
	if !result.Success {
		result.Reason = res.Err.Error()
	}
	return result
}
