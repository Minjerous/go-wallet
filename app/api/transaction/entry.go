package transaction

type Group struct{}

func (g *Group) Base() *BaseApi {
	return &insBase
}
