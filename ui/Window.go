package ui

import (
	"engo.io/engo"
	"engo.io/ecs"
	"image"
	"image/png"
	"os"
	"log"
	"fmt"
	"engo.io/engo/common"
)

type NESVideo struct {
	filename string
	width		int
	height		int
}

type Frame struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

func (n *NESVideo) Preload() {
	err := engo.Files.Load(n.filename)

	if err != nil {
		log.Panic("Preload: "+ err.Error())
	}
}

func (n *NESVideo) Setup(world *ecs.World) {
	world.AddSystem(new(common.RenderSystem))

	frame := Frame{BasicEntity:ecs.NewBasic()}

	frame.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{0, 0},
		Width:    float32(n.width),
		Height:   float32(n.height),
	}

	texture, err := common.LoadedSprite(n.filename)
	if err != nil {
		log.Panic("Setup: Unable to load texture: " + err.Error())
	}

	frame.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{0, 0},
	}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&frame.BasicEntity, &frame.RenderComponent, &frame.SpaceComponent)
		}
	}

}

func (n *NESVideo) Type() string {
	return "kuso-NES"
}

func Run(path string) {

	file,err := os.Open("assets/" + path)

	if err != nil {
		log.Fatalf("Open file error: %v",err)
	}

	log.Println("File loaded.")
	var config image.Config

	config,err = png.DecodeConfig(file)

		if err != nil {
			log.Fatalf("Read png config error: %v",err)
		}
	file.Close()
	log.Println("Config readed.")

	opts := engo.RunOptions{
		Title: fmt.Sprintf("%v - kuso-NES",path),
		Width: config.Width,
		Height: config.Height,
	}

	engo.Run(opts,&NESVideo{path,config.Width,config.Height})
}