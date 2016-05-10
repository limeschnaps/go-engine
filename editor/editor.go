package editor

import (
	"github.com/walesey/go-engine/actor"
	"github.com/walesey/go-engine/assets"
	"github.com/walesey/go-engine/controller"
	"github.com/walesey/go-engine/editor/models"
	"github.com/walesey/go-engine/engine"
	"github.com/walesey/go-engine/renderer"
	"github.com/walesey/go-engine/ui"
	vmath "github.com/walesey/go-engine/vectormath"
)

type Editor struct {
	renderer          renderer.Renderer
	gameEngine        engine.Engine
	currentMap        *editorModels.MapModel
	rootMapNode       *renderer.Node
	customController  *controller.ActionMap
	controllerManager *controller.ControllerManager
	uiAssets          ui.HtmlAssets
	mainMenu          *ui.Window
	overviewMenu      *Overview
	mainMenuOpen      bool
	progressBar       *ui.Window
	fileBrowser       *FileBrowser
	fileBrowserOpen   bool
}

func New() *Editor {
	return &Editor{
		uiAssets:    ui.NewHtmlAssets(),
		rootMapNode: renderer.CreateNode(),
		currentMap: &editorModels.MapModel{
			Name: "default",
			Root: editorModels.NewNodeModel("root"),
		},
	}
}

func (e *Editor) Start() {
	glRenderer := &renderer.OpenglRenderer{
		WindowTitle: "GoEngine editor",
	}
	e.renderer = glRenderer
	e.gameEngine = engine.NewEngine(e.renderer)

	e.gameEngine.Start(func() {
		//lighting
		e.renderer.CreateLight(0.0, 0.0, 0.0, 0.5, 0.5, 0.5, 0.7, 0.7, 0.7, true, vmath.Vector3{0.3, -1, 0.2}, 0)

		//Sky
		skyImg, err := assets.ImportImage("TestAssets/Files/skybox/cubemap.png")
		if err == nil {
			e.gameEngine.Sky(assets.CreateMaterial(skyImg, nil, nil, nil), 999999)
		}

		//root node
		e.gameEngine.AddSpatial(e.rootMapNode)

		//input/controller manager
		e.controllerManager = controller.NewControllerManager(glRenderer.Window)

		//camera + player
		camera := e.gameEngine.Camera()
		freeMoveActor := actor.NewFreeMoveActor(camera)
		freeMoveActor.MoveSpeed = 10.0
		mainController := controller.NewBasicMovementController(freeMoveActor)
		e.controllerManager.AddController(mainController)
		e.gameEngine.AddUpdatable(freeMoveActor)

		//custom controller
		e.customController = controller.NewActionMap()
		e.controllerManager.AddController(e.customController)

		e.setupUI()

		//event loop
		e.gameEngine.AddUpdatable(engine.UpdatableFunc(func(dt float64) {
			//TODO:
		}))
	})
}