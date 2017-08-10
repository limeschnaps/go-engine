package assets

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"image"

	"github.com/limeschnaps/go-engine/renderer"
)

//vericies format : x,y,z,   nx,ny,nz,tx,ty,tz,btx,bty,btz,   u,v,  r,g,b,a
//Indicies format : f1,f2,f3 (triangles)
type objData struct {
	Name     string
	Indicies []uint32
	Vertices []float32
	Mtl      *mtlData
}

type mtlData struct {
	Name                  string
	Ns, Ka, Kd, Ks, Ni, D float32
	Illum                 int

	maps map[string]image.Image
}

//imports an obj from a filePath and return a Geometry
func ImportObj(filePath string) (geometry *renderer.Geometry, material *renderer.Material, err error) {

	obj := &objData{Indicies: make([]uint32, 0, 0), Vertices: make([]float32, 0, 0)}
	vertexList := make([]float32, 0, 0)
	uvList := make([]float32, 0, 0)
	normalList := make([]float32, 0, 0)

	//split the file name from the file path
	filePathTokens := strings.Split(strings.Replace(filePath, "\\", "/", -1), "/")
	fileName := filePathTokens[len(filePathTokens)-1]
	path := strings.TrimRight(filePath, fileName)

	//open the file and read all lines
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening geometry file: %v\n", err)
		return
	}
	defer file.Close()

	var mtlErr error
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) > 0 {
			dataType := tokens[0]
			if dataType == "o" { //sub mesh
				obj.Name = tokens[1]
			} else if dataType == "v" { //xyz vertex
				vertexList = append(vertexList, stf(tokens[1]), stf(tokens[2]), stf(tokens[3]))
			} else if dataType == "vt" { //uv coord
				uvList = append(uvList, stf(tokens[1]), stf(tokens[2]))
			} else if dataType == "vn" { //xyz vertex normal
				normalList = append(normalList, stf(tokens[1]), stf(tokens[2]), stf(tokens[3]))
			} else if dataType == "f" { // v/t/n face
				obj.processFace(line, vertexList, uvList, normalList)
			} else if dataType == "mtllib" {
				obj.Mtl, mtlErr = importMTL(path, tokens[1])
			} else if dataType == "usemtl" { //mtl material
				//TODO: multiple mtls
			}
		}
	}

	geometry = renderer.CreateGeometry(obj.Indicies, obj.Vertices)
	if mtlErr == nil && obj.Mtl != nil {
		textures := []*renderer.Texture{}
		for key, img := range obj.Mtl.maps {
			textures = append(textures, renderer.NewTexture(key, img, true))
		}
		material = renderer.NewMaterial(textures...)
	}

	if err = scanner.Err(); err != nil {
		fmt.Printf("Error loading geometry: %v\n", err)
		return
	}
	return
}

//returns corresponding index (0,1,2...)
func (obj *objData) pushVert(x, y, z, nx, ny, nz, u, v, r, g, b, a float32) uint32 {
	obj.Vertices = append(obj.Vertices, x, y, z, nx, ny, nz, u, v, r, g, b, a)
	return (uint32)((len(obj.Vertices) / 12) - 1)
}

func (obj *objData) pushIndex(indicies ...uint32) {
	obj.Indicies = append(obj.Indicies, indicies...)
}

//parses a single triangle vertex, returning the newly generated index
func (obj *objData) processFaceVertex(token string, vertexList, uvList, normalList []float32) uint32 {
	face := strings.Split(token, "/")
	var index int32

	//vertex
	vx := (float32)(0.0)
	vy := (float32)(0.0)
	vz := (float32)(0.0)
	if len(face) > 0 && face[0] != "" {
		index = (sti(face[0]) - 1) * 3
		vx = vertexList[index]
		vy = vertexList[index+1]
		vz = vertexList[index+2]
	}

	//texture
	vtx := (float32)(0.0)
	vty := (float32)(0.0)
	if len(face) > 1 && face[1] != "" {
		index = (sti(face[1]) - 1) * 2
		vtx = uvList[index]
		vty = uvList[index+1]
	}

	//normal / tangents
	nx := (float32)(0.0)
	ny := (float32)(0.0)
	nz := (float32)(0.0)
	if len(face) > 2 && face[2] != "" {
		index = (sti(face[2]) - 1) * 3
		nx = normalList[index]
		ny = normalList[index+1]
		nz = normalList[index+2]
	}

	return obj.pushVert(vx, vy, vz, nx, ny, nz, vtx, vty, 1.0, 1.0, 1.0, 1.0)
}

//Processes a polygonal face by splitting it into triangles
func (obj *objData) processFace(line string, vertexList, uvList, normalList []float32) {
	tokens := strings.Fields(line)
	if tokens[0] == "f" {
		tokens = append(tokens[:0], tokens[1:]...)
		for len(tokens) > 0 {
			tempTokens := make([]string, 0, 0)
			for i := 0; i < (len(tokens) - 1); i += 2 {
				obj.pushIndex(obj.processFaceVertex(tokens[i], vertexList, uvList, normalList))
				obj.pushIndex(obj.processFaceVertex(tokens[i+1], vertexList, uvList, normalList))
				if len(tokens) > (i + 2) {
					obj.pushIndex(obj.processFaceVertex(tokens[i+2], vertexList, uvList, normalList))
				} else {
					obj.pushIndex(obj.processFaceVertex(tokens[0], vertexList, uvList, normalList))
				}
				if len(tokens) > 4 {
					tempTokens = append(tempTokens, tokens[i])
				}
			}
			if len(tokens) > 4 && len(tokens)%2 == 1 {
				tempTokens = append(tempTokens, tokens[len(tokens)-1])
			}
			tokens = tempTokens
		}
	}
}

//Returns mtl object data type
func importMTL(filePath, fileName string) (*mtlData, error) {
	mtl := &mtlData{maps: make(map[string]image.Image)}

	file, err := os.Open(filePath + fileName)
	if err != nil {
		fmt.Printf("Error opening material file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) > 0 {
			dataType := tokens[0]
			var err error = nil
			switch dataType {
			case "newmtl":
				mtl.Name = tokens[1]
			case "Ns":
				mtl.Ns = stf(tokens[1])
			//TODO: Other mtl variables
			case "map_Kd":
				mtl.maps["diffuseMap"], err = ImportImage(filePath + tokens[1])
			case "map_Spec":
				mtl.maps["specularMap"], err = ImportImage(filePath + tokens[1])
			case "map_AO":
				mtl.maps["aoMap"], err = ImportImage(filePath + tokens[1])
			case "map_Disp":
				mtl.maps["normalMap"], err = ImportImage(filePath + tokens[1])
			case "map_Roughness":
				mtl.maps["roughnessMap"], err = ImportImage(filePath + tokens[1])
			case "map_Metalness":
				mtl.maps["metalnessMap"], err = ImportImage(filePath + tokens[1])
			case "map_Composite":
				mtl.maps["compositeMap"], err = ImportImage(filePath + tokens[1])
			case "map_Glow":
				mtl.maps["glowMap"], err = ImportImage(filePath + tokens[1])
			}
			if err != nil {
				log.Printf("Error parsing mtl data %v: %v\n", dataType, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error loading geometry file: %v\n", err)
		return nil, err
	}

	return mtl, nil
}

//string to float32
func stf(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		log.Fatal(err)
	}
	return (float32)(f)
}

//string to int32
func sti(s string) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return (int32)(i)
}
