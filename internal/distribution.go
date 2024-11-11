package internal

import (
	rand "math/rand/v2"
)

type RandomVariable[T any] interface {
	Sample() T
}

type Uniform struct {
	a int64
	b int64
	r *rand.Rand
}

type Normal struct {
	mean   float64
	stddev float64
	r      *rand.Rand
}

type Poisson struct {
	lambda int64
	r      *rand.Rand
}

type Bernoulli struct {
	p float64
	r *rand.Rand
}

func NewUniform(a int64, b int64, r *rand.Rand) *Uniform {
	d := &Uniform{
		a: a,
		b: b,
		r: r,
	}

	return d
}

func (d *Uniform) Sample() int64 {
	w := d.b - d.a
	return d.a + d.r.Int64N(w)
}

func NewNormal(mean float64, stddev float64, r *rand.Rand) *Normal {
	d := &Normal{
		mean:   mean,
		stddev: stddev,
		r:      r,
	}

	return d
}

func (d *Normal) Sample() int64 {
	return int64(d.r.NormFloat64()*d.stddev + d.mean)
}

func NewBernoulli(p float64, r *rand.Rand) *Bernoulli {
	d := &Bernoulli{
		p: p,
		r: r,
	}

	return d
}

func (d *Bernoulli) Sample() bool {
	return d.r.Float64() < d.p
}
