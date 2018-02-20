package scene

import (
	"github.com/Shnifer/flierproto1/texture"
	"github.com/veandco/go-sdl2/sdl"
)

type StaticImage struct {
	scene   *BScene
	texName string
	Tex     *sdl.Texture
	ZLayer  ZLayer
}

func NewStaticImage(texName string, ZLayer ZLayer) *StaticImage {
	return &StaticImage{texName: texName, ZLayer: ZLayer}
}

func (si *StaticImage) GetID() string {
	return ""
}

func (si *StaticImage) Init(scene *BScene) {
	si.scene = scene
	si.Tex = texture.Cache.GetTexture(si.texName)
}

func (si *StaticImage) Update(dt float32) {
	//nothing
}

func (si *StaticImage) Draw(r *sdl.Renderer) RenderReqList {
	//На весь экран
	//TODO: определить порядок ФОн - объекты - Интерфейс
	return RenderReqList{NewRenderReqSimple(si.Tex, nil, nil, si.ZLayer)}
}
