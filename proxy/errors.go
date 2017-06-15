package proxy

import "errors"

var errWritePacket = errors.New("error while writing packet payload")
var errNoQueryPacket = errors.New("malformed packet")
var errInvalidProxyParams = errors.New("both proxy and mysql hosts must be set")
