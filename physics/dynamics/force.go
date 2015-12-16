package dynamics

import (
	vmath "github.com/walesey/go-engine/vectormath"
)

type ForceStore struct {
	forces map[string]Force
}

type Force interface {
	DoStep(dt float64, phyObj *PhysicsObject)
}

type LinearForce struct {
	Value vmath.Vector3
}

//acceleration due to gravity
type GravityForce struct {
	Value vmath.Vector3
}

type TorqueForce struct {
	Value vmath.Quaternion
}

// force applied to a fixed local point on the object
type PointForce struct {
	Value, Position vmath.Vector3
	positionLength  float64
}

func NewForceStore() *ForceStore {
	return &ForceStore{make(map[string]Force)}
}

func (fs *ForceStore) GetForce(name string) Force {
	return fs.forces[name]
}

func (fs *ForceStore) AddForce(name string, force Force) *ForceStore {
	fs.forces[name] = force
	return fs
}

func (fs *ForceStore) RemoveForce(name string) *ForceStore {
	delete(fs.forces, name)
	return fs
}

func (fs *ForceStore) DoStep(dt float64, phyObj *PhysicsObject) {
	for _, force := range fs.forces {
		force.DoStep(dt, phyObj)
	}
}

func (force *LinearForce) DoStep(dt float64, phyObj *PhysicsObject) {
	phyObj.Velocity = phyObj.Velocity.Add(
		force.Value.DivideScalar(phyObj.Mass).
			MultiplyScalar(dt))
}

func (force *GravityForce) DoStep(dt float64, phyObj *PhysicsObject) {
	phyObj.Velocity = phyObj.Velocity.Add(
		force.Value.MultiplyScalar(dt))
}

func (force *TorqueForce) DoStep(dt float64, phyObj *PhysicsObject) {
	//TODO:
}

func (force *PointForce) DoStep(dt float64, phyObj *PhysicsObject) {
	value := phyObj.Orientation.Apply(force.Value)
	if force.Position.ApproxEqual(vmath.Vector3{0, 0, 0}, 0.00001) {
		//linear only
		phyObj.Velocity = phyObj.Velocity.Add(value.DivideScalar(phyObj.Mass).MultiplyScalar(dt))
	} else {
		//update position length
		if !vmath.ApproxEqual(force.Position.LengthSquared(), force.positionLength, 0.001) {
			force.positionLength = force.Position.Length()
		}

		// linear
		position := phyObj.Orientation.Apply(force.Position)
		positionNorm := position.DivideScalar(force.positionLength).MultiplyScalar(-1)
		linearForce := positionNorm.MultiplyScalar(value.Dot(positionNorm))
		phyObj.Velocity = phyObj.Velocity.Add(linearForce.DivideScalar(phyObj.Mass).MultiplyScalar(dt))

		//angular
		angV := phyObj.AngularVelocityVector()
		torque := value.Cross(position)
		newAngV := angV.Add(phyObj.InertiaTensor().Inverse().Transform(torque))
		phyObj.SetAngularVelocityVector(newAngV)
	}
}