package rqres

import (
	"errors"

	"github.com/vitermakov/otusgo-final/pkg/utils/errx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
Преобразование err->*status.Status именно в указанном виде применяться не будет, сделано для упращения обработки.
*/

func FromError(err error) *status.Status {
	logErr := errx.Logic{}
	if errors.As(err, &logErr) {
		return status.Newf(codes.InvalidArgument, "[%d] %s", logErr.Code(), logErr.Error())
	}
	nfErr := errx.NotFound{}
	if errors.As(err, &nfErr) {
		return status.New(codes.NotFound, nfErr.Error())
	}
	base := errx.Base{}
	if errors.As(err, &base) {
		switch base.Kind() {
		case errx.TypePerms:
			return status.New(codes.PermissionDenied, base.Error())
		case errx.TypeFatal:
			return status.New(codes.Internal, base.Error())
		}
	}
	return status.New(codes.InvalidArgument, err.Error())
}
