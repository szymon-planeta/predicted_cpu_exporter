package algorithm

type Algorithm interface {
	StoreData(float64)
	Predict() float64
}

type Arma struct {

}

func (a *Arma) StoreData(data float64) {

}

func (a *Arma) Predict() float64 {
	return 100
}

func NewArma() *Arma  {
	return &Arma {}
}
