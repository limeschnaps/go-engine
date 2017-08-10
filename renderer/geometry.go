package renderer

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/limeschnaps/go-engine/util"
)

const VertexStride = 12

//Geometry
type Geometry struct {
	VboId, IboId   uint32
	Loaded         bool
	VboDirty       bool
	boundingRadius float32
	Indicies       []uint32
	Verticies      []float32
}

//vericies format : x,y,z,   nx,ny,nz,   u,v,  r,g,b,a
//indicies format : f1,f2,f3 (triangles)
func CreateGeometry(indicies []uint32, verticies []float32) *Geometry {
	return &Geometry{
		Indicies:  indicies,
		Verticies: verticies,
	}
}

func (geometry *Geometry) Copy() *Geometry {
	indicies := make([]uint32, len(geometry.Indicies))
	verticies := make([]float32, len(geometry.Verticies))
	copy(indicies, geometry.Indicies)
	copy(verticies, geometry.Verticies)
	return CreateGeometry(indicies, verticies)
}

func (geometry *Geometry) Draw(renderer Renderer, transform mgl32.Mat4) {
	if len(geometry.Verticies) == 0 && len(geometry.Indicies) == 0 {
		geometry.boundingRadius = 0
		return
	}
	if !geometry.Loaded || geometry.VboDirty {
		geometry.boundingRadius = geometry.MaximalPointFromGeometry().Len()
	}
	renderer.DrawGeometry(geometry, transform)
}

func (geometry *Geometry) Destroy(renderer Renderer) {
	renderer.DestroyGeometry(geometry)
	geometry.Loaded = false
}

func (geometry *Geometry) Center() mgl32.Vec3 {
	return mgl32.Vec3{0, 0, 0}
}

func (geometry *Geometry) SetParent(parent *Node) {}

func (geometry *Geometry) ClearBuffers() {
	geometry.Indicies = geometry.Indicies[:0]
	geometry.Verticies = geometry.Verticies[:0]
}

func (geometry *Geometry) SetColor(color color.Color) {
	for i := 8; i < len(geometry.Verticies); i = i + VertexStride {
		r, g, b, a := color.RGBA()
		geometry.Verticies[i] = float32(r) / 65535.0
		geometry.Verticies[i+1] = float32(g) / 65535.0
		geometry.Verticies[i+2] = float32(b) / 65535.0
		geometry.Verticies[i+3] = float32(a) / 65535.0
	}
	geometry.updateGeometry()
}

func (geometry *Geometry) Transform(transform mgl32.Mat4) {
	geometry.transformRange(transform, 0)
}

func (geometry *Geometry) transformRange(transform mgl32.Mat4, from int) {
	for i := from; i < len(geometry.Verticies); i = i + VertexStride {
		v := mgl32.TransformCoordinate(mgl32.Vec3{
			geometry.Verticies[i],
			geometry.Verticies[i+1],
			geometry.Verticies[i+2],
		}, transform)
		n := mgl32.TransformNormal(mgl32.Vec3{
			geometry.Verticies[i+3],
			geometry.Verticies[i+4],
			geometry.Verticies[i+5],
		}, transform)
		geometry.Verticies[i] = v.X()
		geometry.Verticies[i+1] = v.Y()
		geometry.Verticies[i+2] = v.Z()
		geometry.Verticies[i+3] = n.X()
		geometry.Verticies[i+4] = n.Y()
		geometry.Verticies[i+5] = n.Z()
	}
	geometry.updateGeometry()
}

//load the verts/indicies of geometry into destination Geometry
func (geometry *Geometry) Optimize(destination *Geometry, transform mgl32.Mat4) {
	vertOffset := len(destination.Verticies)
	indexOffset := len(destination.Indicies)
	destination.Verticies = append(destination.Verticies, geometry.Verticies...)
	destination.Indicies = append(destination.Indicies, geometry.Indicies...)
	for i, _ := range geometry.Indicies {
		destination.Indicies[i+indexOffset] = geometry.Indicies[i] + uint32(vertOffset/VertexStride)
	}
	destination.transformRange(transform, vertOffset)
	geometry.updateGeometry()
}

func (geometry *Geometry) BoundingRadius() float32 {
	return geometry.boundingRadius
}

func (geometry *Geometry) OrthoOrder() int {
	return 0
}

func (geometry *Geometry) SetUVs(uvs ...float32) {
	for i := 0; i < len(uvs); i = i + 2 {
		geometry.Verticies[((i/2)*VertexStride)+6] = uvs[i]
		geometry.Verticies[((i/2)*VertexStride)+7] = uvs[i+1]
	}
	geometry.updateGeometry()
}

func (geometry *Geometry) updateGeometry() {
	geometry.boundingRadius = geometry.MaximalPointFromGeometry().Len()
	geometry.VboDirty = true
}

// Gets the furthest point from the origin in this geometry
func (geometry *Geometry) MaximalPointFromGeometry() mgl32.Vec3 {
	var longestSq float32
	var longest *mgl32.Vec3
	for i := 0; i < len(geometry.Indicies); i = i + 1 {
		index := geometry.Indicies[i]
		v := mgl32.Vec3{
			geometry.Verticies[index*VertexStride],
			geometry.Verticies[index*VertexStride+1],
			geometry.Verticies[index*VertexStride+2],
		}
		if longest == nil || util.Vec3LenSq(v) > longestSq {
			longestSq = util.Vec3LenSq(v)
			longest = &v
		}
	}
	if longest != nil {
		return *longest
	}
	return mgl32.Vec3{}
}

func (geometry *Geometry) RayIntersect(start, direction mgl32.Vec3) (point mgl32.Vec3, ok bool) {
	indicies := geometry.Indicies
	verts := geometry.Verticies
	var distSq float32
	var closest *mgl32.Vec3
	for i := 0; i < len(geometry.Indicies); i = i + 3 {
		indexA, indexB, indexC := indicies[i], indicies[i+1], indicies[i+2]
		a := mgl32.Vec3{verts[indexA*VertexStride], verts[indexA*VertexStride+1], verts[indexA*VertexStride+2]}
		b := mgl32.Vec3{verts[indexB*VertexStride], verts[indexB*VertexStride+1], verts[indexB*VertexStride+2]}
		c := mgl32.Vec3{verts[indexC*VertexStride], verts[indexC*VertexStride+1], verts[indexC*VertexStride+2]}
		if point, intersectOk := util.RayTriangleIntersect(a, b, c, start, direction); intersectOk {
			newDist := util.Vec3LenSq(point.Sub(start))
			if closest == nil || newDist < distSq {
				distSq = newDist
				closest = &point
			}
		}
	}
	if closest != nil {
		point = *closest
		ok = true
	}
	return
}

//Primitives
func CreateBox(width, height float32) *Geometry {
	return CreateBoxWithOffset(width, height, -width/2, -height/2)
}

func CreateBoxWithOffset(width, height, offsetX, offsetY float32) *Geometry {
	verticies := []float32{
		//    x,                y, z, nx, ny, nz, u, v,   r,   g,   b,   a
		offsetX, height + offsetY, 0,  0,  1,  0, 0, 0, 1.0, 1.0, 1.0, 1.0,
		width + offsetX, height + offsetY, 0, 0, 1, 0, 1, 0, 1.0, 1.0, 1.0, 1.0,
		width + offsetX, offsetY, 0, 0, 1, 0, 1, 1, 1.0, 1.0, 1.0, 1.0,
		offsetX, offsetY, 0, 0, 1, 0, 0, 1, 1.0, 1.0, 1.0, 1.0,
	}
	indicies := []uint32{0, 1, 2, 2, 3, 0}
	return CreateGeometry(indicies, verticies)
}

func CreateSkyBox() *Geometry {
	return CreateGeometry(cubeIndicies, skyboxVerticies)
}

func CreateCube() *Geometry {
	geo := CreateBoxWithOffset(1, 1, -0.5, -0.5)
	geo.Transform(mgl32.Translate3D(0, 0, -0.5))
	geo2 := CreateBoxWithOffset(1, 1, -0.5, -0.5)
	geo2.Optimize(geo, mgl32.Translate3D(0, 0, 0.5))
	geo.Indicies = append(geo.Indicies, 0, 1, 4, 4, 5, 0) //top
	geo.Indicies = append(geo.Indicies, 1, 2, 7, 7, 4, 1) //side
	geo.Indicies = append(geo.Indicies, 2, 3, 6, 6, 7, 2) //bottom
	geo.Indicies = append(geo.Indicies, 3, 0, 5, 5, 6, 3) //side
	return geo
}

// CreateBeam - creates a square prism oriented along the vector
func CreateBeam(width float32, vector mgl32.Vec3) *Geometry {
	geo := CreateCube()
	geo.Transform(AlignedOrientation(width, vector))
	return geo
}

// CreateBeam - creates a square prism oriented along the vector
func AlignedOrientation(width float32, vector mgl32.Vec3) mgl32.Mat4 {
	len := vector.Len()
	return mgl32.QuatBetweenVectors(mgl32.Vec3{0, 1, 0}, vector).Mat4().
		Mul4(mgl32.Translate3D(0, 0.5*len, 0)).
		Mul4(mgl32.Scale3D(width, len, width))
}
