package vectormath

import "math"

type Vector3 struct {
	X, Y, Z float64
}

type Vector interface {
	Length() float64
	LengthSquared() float64
}

func (v Vector3) LengthSquared() float64 {
	return (v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z)
}

func (v Vector3) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vector3) Normalize() Vector3 {
	if v.LengthSquared() > 0.99999 && v.LengthSquared() < 1.00001 {
		return v
	}
	return v.DivideScalar(v.Length())
}

func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		v.X + other.X,
		v.Y + other.Y,
		v.Z + other.Z,
	}
}

func (v Vector3) Subtract(other Vector3) Vector3 {
	return Vector3{
		v.X - other.X,
		v.Y - other.Y,
		v.Z - other.Z,
	}
}

func (v Vector3) Multiply(other Vector3) Vector3 {
	return Vector3{
		v.X * other.X,
		v.Y * other.Y,
		v.Z * other.Z,
	}
}

func (v Vector3) MultiplyScalar(scalar float64) Vector3 {
	return Vector3{
		v.X * scalar,
		v.Y * scalar,
		v.Z * scalar,
	}
}

func (v Vector3) Divide(other Vector3) Vector3 {
	return Vector3{
		v.X / other.X,
		v.Y / other.Y,
		v.Z / other.Z,
	}
}

func (v Vector3) DivideScalar(scalar float64) Vector3 {
	return Vector3{
		v.X / scalar,
		v.Y / scalar,
		v.Z / scalar,
	}
}

func (v Vector3) Project(other Vector3) Vector3 {
	other = other.Normalize()
	projection := v.Dot(other)
	return other.MultiplyScalar(projection)
}

func (v Vector3) Cross(other Vector3) Vector3 {
	return Vector3{
		(other.Y * v.Z) - (other.Z * v.Y),
		(other.Z * v.X) - (other.X * v.Z),
		(other.X * v.Y) - (other.Y * v.X),
	}
}

func (v Vector3) Lerp(dest Vector3, amount float64) Vector3 {
	return v.MultiplyScalar(1.0 - amount).Add(dest.MultiplyScalar(amount))
}

func (v Vector3) Dot(other Vector3) float64 {
	return (v.X * other.X) + (v.Y * other.Y) + (v.Z * other.Z)
}

func (v Vector3) AngleBetween(other Vector3) float64 {
	return math.Acos(v.Normalize().Dot(other.Normalize()))
}

func (v Vector3) ApproxEqual(other Vector3, epsilon float64) bool {
	return ApproxEqual(v.X, other.X, epsilon) && ApproxEqual(v.Y, other.Y, epsilon) && ApproxEqual(v.Z, other.Z, epsilon)
}