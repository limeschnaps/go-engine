package editor

import (
	"bytes"

	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/limeschnaps/go-engine/actor"
	"github.com/limeschnaps/go-engine/assets"
	"github.com/limeschnaps/go-engine/controller"
	"github.com/limeschnaps/go-engine/editor/models"
	"github.com/limeschnaps/go-engine/effects"
	"github.com/limeschnaps/go-engine/engine"
	"github.com/limeschnaps/go-engine/glfwController"
	"github.com/limeschnaps/go-engine/opengl"
	"github.com/limeschnaps/go-engine/renderer"
	"github.com/limeschnaps/go-engine/ui"
	"github.com/limeschnaps/go-engine/util"
)

type Editor struct {
	assetDir              string
	renderer              renderer.Renderer
	gameEngine            engine.Engine
	currentMap            *editorModels.MapModel
	rootMapNode           *renderer.Node
	customController      controller.Controller
	controllerManager     *glfwController.ControllerManager
	uiAssets              ui.HtmlAssets
	mainMenu              *ui.Window
	mainMenuController    glfwController.Controller
	overviewMenu          *Overview
	mainMenuOpen          bool
	progressBar           *ui.Window
	fileBrowser           *FileBrowser
	fileBrowserOpen       bool
	fileBrowserController glfwController.Controller
	mouseMode             string
	selectSprite          *effects.Sprite
}

func New(assetDir string) *Editor {
	return &Editor{
		assetDir:    assetDir,
		uiAssets:    ui.NewHtmlAssets(),
		rootMapNode: renderer.NewNode(),
		currentMap: &editorModels.MapModel{
			Name: "default",
			Root: editorModels.NewNodeModel("root"),
		},
		mouseMode: "scale",
	}
}

func (e *Editor) Start() {

	glRenderer := opengl.NewOpenglRenderer("GoEngine Editor", 0, 0, false)
	e.renderer = glRenderer
	e.gameEngine = engine.NewEngine(e.renderer)

	e.gameEngine.Start(func() {

		if shader, err := assets.ImportShader(
			filepath.Join(e.assetDir, "shaders/build/pbrComposite.vert"),
			filepath.Join(e.assetDir, "shaders/build/pbrComposite.frag"),
		); err == nil {
			shader.AddTexture("diffuseMap")
			shader.AddTexture("normalMap")
			shader.AddTexture("compositeMap")
			shader.AddTexture("glowMap")
			shader.AddTexture("environmentMap")
			shader.FragDataLocations = []string{"outputColor", "brightColor"}
			e.gameEngine.DefaultShader(shader)
		} else {
			panic("error importing shader")
		}

		if shader, err := assets.ImportShader(
			filepath.Join(e.assetDir, "shaders/build/postEffects/glow.vert"),
			filepath.Join(e.assetDir, "shaders/build/postEffects/glow.frag"),
		); err == nil {
			shader.InputBuffers = 2
			glRenderer.CreatePostEffect(shader)
		} else {
			panic("error importing shaders")
		}

		// Sky cubemap
		skyImg, err := assets.ImportImage(filepath.Join(e.assetDir, "TestAssets/cloudSky.jpg"))
		if err == nil {
			geom := renderer.CreateSkyBox()
			geom.Transform(mgl32.Scale3D(10000, 10000, 10000))
			skyNode := renderer.NewNode()
			skyNode.SetOrientation(mgl32.QuatRotate(1.57, mgl32.Vec3{0, 1, 0}))
			skyNode.Material = renderer.NewMaterial(renderer.NewTexture("diffuseMap", skyImg, false))
			skyNode.RendererParams = renderer.NewRendererParams()
			skyNode.RendererParams.CullBackface = false
			skyNode.RendererParams.Unlit = true
			skyNode.Add(geom)
			e.gameEngine.AddSpatial(skyNode)
			// create an environmentMap using the skybox texture
			envCubeMap := renderer.NewCubemap(skyImg, true)
			e.gameEngine.DefaultCubeMap(envCubeMap)
		}

		l := renderer.NewLight(renderer.DIRECTIONAL)
		l.Color = [3]float32{0.3, 0.3, 0.4}
		e.gameEngine.AddLight(l)

		//root node
		e.gameEngine.AddSpatial(e.rootMapNode)

		//input/controller manager
		e.controllerManager = glfwController.NewControllerManager(glRenderer.Window)

		//camera + player
		camera := e.gameEngine.Camera()
		freeMoveActor := actor.NewFreeMoveActor(camera)
		freeMoveActor.MoveSpeed = 20.0
		freeMoveActor.LookSpeed = 0.002
		mainController := controller.NewBasicMovementController(freeMoveActor, true)
		e.controllerManager.AddController(mainController.(glfwController.Controller))
		e.gameEngine.AddUpdatable(freeMoveActor)

		e.initSelectSprite()
		e.gameEngine.AddUpdatable(engine.UpdatableFunc(e.updateSelectSprite))

		//editor controller
		e.controllerManager.AddController(NewEditorController(e).(glfwController.Controller))

		//custom controller
		e.customController = controller.CreateController()
		e.controllerManager.AddController(e.customController.(glfwController.Controller))

		e.setupUI()
	})
}

func (e *Editor) initSelectSprite() {
	img, _ := assets.DecodeImage(bytes.NewBuffer(util.Base64ToBytes(GeometryIconData)))
	mat := renderer.NewMaterial(renderer.NewTexture("diffuseMap", img, false))
	selectSprite := effects.CreateSprite(1, 1, 1, mat)
	spriteNode := renderer.NewNode()
	spriteNode.RendererParams = &renderer.RendererParams{Unlit: true}
	spriteNode.Add(selectSprite)
	e.selectSprite = selectSprite
	e.gameEngine.AddSpatialTransparent(spriteNode)
}

func (e *Editor) updateSelectSprite(dt float64) {
	selectedModel, _ := e.overviewMenu.getSelectedNode(e.currentMap.Root)
	if selectedModel == nil {
		e.selectSprite.SetScale(mgl32.Vec3{})
	} else {
		camPos := e.gameEngine.Camera().Translation
		e.selectSprite.SetCameraLocation(camPos)
		e.selectSprite.Update(dt)
		size := selectedModel.Translation.Sub(camPos).Len() * 0.02
		e.selectSprite.SetScale(mgl32.Vec3{size, size, size})
		translation, err := e.rootMapNode.RelativePosition(selectedModel.GetNode())
		e.selectSprite.SetTranslation(translation)
		if err != nil {
			e.selectSprite.SetScale(mgl32.Vec3{})
		}
	}
}
