package user

type Group struct{}

func (g *Group) Sign() *SSign {
	return &insSign
}

func (g *Group) Info() *SInfo {
	return &insInfo
}
