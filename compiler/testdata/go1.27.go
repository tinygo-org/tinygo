package main

// genericMethod has both a regular method and a "generic method": a method
// with its own type parameter, independent of any type parameter on the
// receiver. This is a Go 1.27 feature (see math/rand/v2.Rand.N for a
// real-world example). Boxing such a value into an interface must not try to
// encode the generic method's type-parameterized signature into a type code.
type genericMethod struct{}

func (t genericMethod) Regular(n int) int { return n }

func (t genericMethod) GenericParam[X int | int64](n X) X { return n }

// onlyGenericMethod's only method is generic, so its runtime type must have
// an empty method set (hasMethodSet must become false, not just numMethods).
type onlyGenericMethod struct{}

func (t onlyGenericMethod) GenericParam[X int | int64](n X) X { return n }

func useGenericMethods() (any, any) {
	var g genericMethod
	var o onlyGenericMethod
	return any(g), any(o)
}
