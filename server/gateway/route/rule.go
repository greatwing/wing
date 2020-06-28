//路由规则
package route

import "sync"

var (
	ruleByMsgID      = map[int]*Rule{}
	ruleByMsgIDGuard sync.RWMutex
)

type Rule struct {
	MsgId   int
	SvcName string
}

//添加路由规则
func AddRule(msgId int, svcName string) {
	ruleByMsgIDGuard.Lock()
	ruleByMsgID[msgId] = &Rule{
		MsgId:   msgId,
		SvcName: svcName,
	}
	ruleByMsgIDGuard.Unlock()
}

func AddRules(msgIdBegin, msgIdEnd int, svcName string) {
	for i := msgIdBegin; i <= msgIdEnd; i++ {
		AddRule(i, svcName)
	}
}

func AddRulesByArray(msgIdArray []int, svcName string) {
	for id := range msgIdArray {
		AddRule(id, svcName)
	}
}

//获取路由规则
func GetRuleByMsgID(msgId int) *Rule {
	ruleByMsgIDGuard.RLock()
	defer ruleByMsgIDGuard.RUnlock()

	if rule, ok := ruleByMsgID[msgId]; ok {
		return rule
	}

	return nil
}
