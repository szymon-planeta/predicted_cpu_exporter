package algorithm

type Algorithm interface {
	StoreData(float64)
	Predict() float64
}

type Arma struct {
	data []float64
}

func (a *Arma) StoreData(n float64) {
	if len(a.data) < 3 {
		a.data = []float64{n, n, n}
	} else {
		a.data = append(a.data[:0], a.data[1:]...)
		a.data = append(a.data[:], n)
	}
}

func (a *Arma) Predict() float64 {
	beta := 0.8
	gamma := 0.15
	return (beta*a.data[2]) + (gamma*a.data[1]) + ((1-(beta+gamma))*a.data[0])
}

func NewArma() *Arma  {
	return &Arma {}
}
