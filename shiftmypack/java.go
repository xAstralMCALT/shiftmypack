package shiftmypack

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	"github.com/xAstralMCALT/shiftmypack/shiftmypack/bedrock"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/image"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/internal/fsutil"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/java"
)

// bedrockReplacer replaces the texture names with the correct names.
var bedrockReplacer = strings.NewReplacer(
	"chainmail_layer_1", "chain_1",
	"chainmail_layer_2", "chain_2",

	"diamond_layer_1", "diamond_1",
	"diamond_layer_2", "diamond_2",

	"gold_layer_1", "gold_1",
	"gold_layer_2", "gold_2",

	"iron_layer_1", "iron_1",
	"iron_layer_2", "iron_2",

	"leather_layer_1", "leather_1",
	"leather_layer_2", "leather_2",

	"netherite_layer_1", "netherite_1",
	"netherite_layer_2", "netherite_2",
)

func PortJavaEditionPackAndExtract(pack java.ResourcePack, outputDirector string) error {
	newPack := bedrock.ResourcePack{}
	newPack.Name = pack.Name

	pack.PackIcon.Name = "pack_icon.png"
	newPack.PackIcon = pack.PackIcon

	newPack.Icons = pack.Icons
	newPack.Particles = pack.Particles

	newPack.Items = pack.Items
	newPack.Blocks = pack.Blocks

	var newArmors []image.Texture
	for _, a := range pack.Armors {
		a.Name = bedrockReplacer.Replace(a.Name)
		newArmors = append(newArmors, a)
	}
	newPack.Armors = newArmors
	newPack.CubeMaps, _ = bedrock.CubemapsFromTexture(pack.Skies[0])

	tmp := "shiftmypack-" + strconv.Itoa(int(rand.IntN(99999)))
	tmpPath := "tmp/" + tmp + ".mcpack"

	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		return err
	}
	if err := newPack.WriteZip(tmpPath); err != nil {
		return err
	}
	fmt.Println("mcpack file written to:", tmpPath)

	output := outputDirector + "/" + tmp
	if err := fsutil.Unzip(tmpPath, output); err != nil {
		return err
	}
	fmt.Println("extracted mcpack to:", output)
	return nil
}

func PortJavaEditionPack(pack java.ResourcePack, output string) error {
	newPack := bedrock.ResourcePack{}
	newPack.Name = pack.Name

	pack.PackIcon.Name = "pack_icon.png"
	newPack.PackIcon = pack.PackIcon

	newPack.Icons = pack.Icons
	newPack.Particles = pack.Particles

	newPack.Items = pack.Items
	newPack.Blocks = pack.Blocks

	var newArmors []image.Texture
	for _, a := range pack.Armors {
		a.Name = bedrockReplacer.Replace(a.Name)
		newArmors = append(newArmors, a)
	}
	newPack.Armors = newArmors
	newPack.CubeMaps, _ = bedrock.CubemapsFromTexture(pack.Skies[0])

	if err := newPack.WriteZip(output); err != nil {
		return err
	}
	fmt.Println("mcpack file written to:", output)
	return nil
}
