package renderer

import (
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/limeschnaps/go-engine/util"
)

type Transparency int

const (
	_ Transparency = iota
	NON_EMISSIVE
	EMISSIVE
)

type RendererParams struct {
	CullBackface bool
	DepthTest    bool
	DepthMask    bool
	Unlit        bool
	Transparency
}

//Node
type Node struct {
	Transform   mgl32.Mat4
	Scale       mgl32.Vec3
	Translation mgl32.Vec3
	Orientation mgl32.Quat

	FrustrumCulling bool
	OrthoOrderValue int
	Shader          *Shader
	Material        *Material
	CubeMap         *CubeMap
	RendererParams  *RendererParams

	parent   *Node
	children []Spatial
	deleted  []Spatial
}

type byOrthoOrder []Spatial

func (b byOrthoOrder) Len() int {
	return len(b)
}

func (b byOrthoOrder) Less(i, j int) bool {
	return b[i].OrthoOrder() < b[j].OrthoOrder()
}

func (b byOrthoOrder) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func NewNode() *Node {
	node := &Node{
		children:        make([]Spatial, 0, 0),
		deleted:         make([]Spatial, 0, 0),
		Translation:     mgl32.Vec3{0, 0, 0},
		Orientation:     mgl32.QuatIdent(),
		FrustrumCulling: false,
	}
	node.SetScale(mgl32.Vec3{1, 1, 1})
	return node
}

func NewRendererParams() *RendererParams {
	defaults := DefaultRendererParams()
	return &defaults
}

func DefaultRendererParams() RendererParams {
	return RendererParams{
		DepthTest:    true,
		DepthMask:    true,
		CullBackface: true,
		Unlit:        false,
		Transparency: NON_EMISSIVE,
	}
}

func (node *Node) Draw(renderer Renderer, transform mgl32.Mat4) {
	sort.Sort(byOrthoOrder(node.children))
	tx := transform.Mul4(node.Transform)
	for _, child := range node.children {
		node.DrawChild(renderer, tx, child)
	}
	node.cleanupDeleted(renderer)
}

func (node *Node) DrawChild(renderer Renderer, transform mgl32.Mat4, child Spatial) {
	if node.FrustrumCulling && !renderer.Camera().CameraContainsSphere(
		renderer.WindowDimensions(),
		child.BoundingRadius()*util.MaxF32(node.Scale[0], node.Scale[1], node.Scale[2]),
		mgl32.TransformCoordinate(child.Center(), transform),
	) {
		return
	}

	node.setRenderStates(renderer)
	child.Draw(renderer, transform)
}

func (node *Node) setRenderStates(renderer Renderer) {
	renderer.UseShader(nil)
	renderer.UseMaterial(nil)
	renderer.UseRendererParams(DefaultRendererParams())

	for parent := node; parent != nil; parent = parent.parent {
		if parent.Shader != nil {
			renderer.UseShader(parent.Shader)
			break
		}
	}
	for parent := node; parent != nil; parent = parent.parent {
		if parent.Material != nil {
			renderer.UseMaterial(parent.Material)
			break
		}
	}
	for parent := node; parent != nil; parent = parent.parent {
		if parent.CubeMap != nil {
			renderer.UseCubeMap(parent.CubeMap)
			break
		}
	}
	for parent := node; parent != nil; parent = parent.parent {
		if parent.RendererParams != nil {
			renderer.UseRendererParams(*parent.RendererParams)
			break
		}
	}
}

func (node *Node) Destroy(renderer Renderer) {
	for _, child := range node.children {
		child.Destroy(renderer)
	}
	node.Material.Destroy(renderer)
	node.CubeMap.Destroy(renderer)
	node.cleanupDeleted(renderer)
}

func (node *Node) cleanupDeleted(renderer Renderer) {
	for _, child := range node.deleted {
		child.Destroy(renderer)
	}
	node.deleted = node.deleted[:0]
}

func (node *Node) Center() mgl32.Vec3 {
	return node.Translation
}

func (node *Node) SetParent(parent *Node) {
	// nodes can only be added to one parent at a time
	if node.parent != nil {
		node.parent.Remove(node, false)
	}
	node.parent = parent
}

func (node *Node) Add(spatial Spatial) {
	spatial.SetParent(node)
	node.children = append(node.children, spatial)
}

func (node *Node) Remove(spatial Spatial, destroy bool) {
	for i, child := range node.children {
		if child == spatial {
			node.children[i] = node.children[len(node.children)-1]
			node.children[len(node.children)-1] = nil
			node.children = node.children[:len(node.children)-1]
			child.SetParent(nil)
			if destroy {
				node.deleted = append(node.deleted, child)
			}
			break
		}
	}
}

func (node *Node) RemoveAll(destroy bool) {
	for _, child := range node.children {
		child.SetParent(nil)
	}
	if destroy {
		node.deleted = append(node.deleted, node.children...)
	}
	node.children = node.children[:0]
}

func (node *Node) SetScale(scale mgl32.Vec3) {
	node.Scale = scale
	node.Transform = util.Mat4From(node.Scale, node.Translation, node.Orientation)
}

func (node *Node) SetTranslation(translation mgl32.Vec3) {
	node.Translation = translation
	node.Transform = util.Mat4From(node.Scale, node.Translation, node.Orientation)
}

func (node *Node) SetOrientation(orientation mgl32.Quat) {
	node.Orientation = orientation
	node.Transform = util.Mat4From(node.Scale, node.Translation, node.Orientation)
}

func (node *Node) SetRotation(angle float32, axis mgl32.Vec3) {
	node.Orientation = mgl32.QuatRotate(angle, axis)
	node.Transform = util.Mat4From(node.Scale, node.Translation, node.Orientation)
}

func (node *Node) OptimizeNode() *Geometry {
	geometry := CreateGeometry(make([]uint32, 0, 0), make([]float32, 0, 0))
	node.Optimize(geometry, node.Transform)
	geometry.VboDirty = true
	return geometry
}

func (node *Node) Optimize(geometry *Geometry, transform mgl32.Mat4) {
	newTransform := transform.Mul4(node.Transform)
	for _, child := range node.children {
		child.Optimize(geometry, newTransform)
	}
}

// RelativePosition - find the position of n relative to node (n must be a child of node)
func (node *Node) RelativePosition(n *Node) (mgl32.Vec3, error) {
	if node == n {
		return mgl32.Vec3{}, nil
	}
	for _, child := range node.children {
		if childNode, ok := child.(*Node); ok {
			if rPost, err := childNode.RelativePosition(n); err == nil {
				return mgl32.TransformCoordinate(rPost, childNode.Transform), nil
			}
		}
	}
	return mgl32.Vec3{}, fmt.Errorf("Node not found")
}

func (node *Node) BoundingRadius() float32 {
	var radius float32
	for _, child := range node.children {
		boundingRadius := child.BoundingRadius()
		if !child.Center().ApproxEqual(mgl32.Vec3{}) {
			boundingRadius += child.Center().Len()
		}
		if boundingRadius > radius {
			radius = boundingRadius
		}
	}

	point := mgl32.TransformCoordinate(mgl32.Vec3{}, node.Transform)
	return mgl32.TransformCoordinate(mgl32.Vec3{radius, 0, 0}, node.Transform).Sub(point).Len()
}

func (node *Node) OrthoOrder() int {
	return node.OrthoOrderValue
}

func (node *Node) SetFrustrumCullingRecursive(enable bool) {
	node.FrustrumCulling = enable
	for _, child := range node.children {
		if childNode, ok := child.(*Node); ok {
			childNode.SetFrustrumCullingRecursive(enable)
		}
	}
}

func (node *Node) RayIntersect(start, direction mgl32.Vec3) (point mgl32.Vec3, ok bool) {
	inverseTx := node.Transform.Inv()
	s := mgl32.TransformCoordinate(start, inverseTx)
	d := mgl32.TransformNormal(direction, inverseTx)
	var distSq float32
	var closest *mgl32.Vec3
	for _, child := range node.children {
		var intersect mgl32.Vec3
		var intersectOk bool
		if childNode, okNode := child.(*Node); okNode {
			intersect, intersectOk = childNode.RayIntersect(s, d)
		}
		if childGeometry, okGeom := child.(*Geometry); okGeom {
			intersect, intersectOk = childGeometry.RayIntersect(s, d)
		}
		if intersectOk {
			newDist := util.Vec3LenSq(intersect.Sub(s))
			if closest == nil || newDist < distSq {
				distSq = newDist
				closest = &intersect
			}
		}
	}
	if closest != nil {
		point = *closest
		ok = true
		point = mgl32.TransformCoordinate(point, node.Transform)
	}
	return
}
