package dto

import (
	"net"

	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
)

func IPNetModel(req *pb.IPNet) (net.IPNet, error) {
	_, mask, err := net.ParseCIDR(req.IPNet)
	if err != nil {
		return net.IPNet{}, err
	}
	return *mask, nil
}
