package assets

import (
	"log"

	"github.com/limeschnaps/go-engine/editor/models"
	"github.com/limeschnaps/go-engine/renderer"
)

type geomImport struct {
	geometry *renderer.Geometry
	material *renderer.Material
	callback func(geometry *renderer.Geometry, material *renderer.Material)
}

type mapImport struct {
	node     *renderer.Node
	model    *editorModels.NodeModel
	callback func(node *renderer.Node, model *editorModels.NodeModel)
}

// The loader allows asyncronous loading of obj and map files
type Loader struct {
	geoms chan geomImport
	maps  chan mapImport
}

func NewLoader() *Loader {
	return &Loader{
		geoms: make(chan geomImport, 256),
		maps:  make(chan mapImport, 256),
	}
}

func (loader *Loader) Update(dt float64) {
	for {
		select {
		case g := <-loader.geoms:
			g.callback(g.geometry, g.material)
		case m := <-loader.maps:
			m.callback(m.node, m.model)
		default:
			return
		}
	}
}

func (loader *Loader) LoadMap(path string, callback func(node *renderer.Node, model *editorModels.NodeModel)) {
	go func() {
		srcModel := LoadMap(path)
		destNode := renderer.NewNode()
		loadedModel := LoadMapToNode(srcModel.Root, destNode)
		loader.maps <- mapImport{
			node:     destNode,
			model:    loadedModel,
			callback: callback,
		}
	}()
}

func (loader *Loader) LoadObj(path string, callback func(geometry *renderer.Geometry, material *renderer.Material)) {
	go func() {
		loadedGeometry, loadedMaterial, err := ImportObjCached(path)
		if err != nil {
			log.Println("Error Loading Obj: ", err)
		} else {
			loader.geoms <- geomImport{
				geometry: loadedGeometry,
				material: loadedMaterial,
				callback: callback,
			}
		}
	}()
}
