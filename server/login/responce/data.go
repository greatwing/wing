package responce

const (
	Succeed = iota
	UidEmpty
	NoAvailableGateway
	GenTokenFailed //生成新token失败
)

type LoginData struct {
	Uid         string   `json:"uid"`
	Token       string   `json:"token"`
	GatewayAddr []string `json:"gateway_addr"`
}

type LoginResponce struct {
	Ret  int        `json:"ret"`
	Msg  string     `json:"msg,omitempty"`
	User *LoginData `json:"data,omitempty"`
}
