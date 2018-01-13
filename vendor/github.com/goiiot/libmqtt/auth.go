package libmqtt

type authPacket struct {
}

func (a *authPacket) Type() CtrlType {
	return CtrlAuth
}
