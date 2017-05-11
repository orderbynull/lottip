package proxy

import "errors"

var ErrWritePacket = errors.New("error while writing packet payload")
var ErrNoQueryPacket = errors.New("malformed packet")
var ErrInvalidProxyParams = errors.New("both proxy and mysql hosts must be set")
