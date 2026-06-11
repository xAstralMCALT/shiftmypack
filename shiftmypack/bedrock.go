package shiftmypack

import (
	"fmt"
	"strings"

	"github.com/xAstralMCALT/shiftmypack/shiftmypack/bedrock"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/image"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/java"
)

// bedrockArmorReplacer reverses the texture names from bedrock format back to java format.
var bedrockArmorReplacer = strings.NewReplacer(
	"chain_1", "chainmail_layer_1",
	"chain_2", "chainmail_layer_2",

	"diamond_1", "diamond_layer_1",
	"diamond_2", "diamond_layer_2",

	"gold_1", "gold_layer_1",
	"gold_2", "gold_layer_2",

	"iron_1", "iron_layer_1",
	"iron_2", "iron_layer_2",

	"leather_1", "leather_layer_1",
	"leather_2", "leather_layer_2",

	"netherite_1", "netherite_layer_1",
	"netherite_2", "netherite_layer_2",
)

// PortBedrockPack converts a Bedrock Edition pack to Java Edition format.
func PortBedrockPack(pack bedrock.ResourcePack, output string) error {
	return PortBedrockPackWithVersion(pack, output, "1.20")
}

// PortBedrockPackWithVersion converts a Bedrock Edition pack to Java Edition format with specified version.
func PortBedrockPackWithVersion(pack bedrock.ResourcePack, output string, version string) error {
	newPack := java.ResourcePack{}
	newPack.Name = pack.Name

	pack.PackIcon.Name = "pack.png"
	newPack.PackIcon = pack.PackIcon

	newPack.Icons = pack.Icons
	newPack.Particles = pack.Particles

	newPack.Items = pack.Items
	newPack.Blocks = pack.Blocks

	// Convert cubemaps to skies
	if len(pack.CubeMaps) > 0 {
		newPack.Skies = []image.Texture{pack.CubeMaps[0]}
	}

	// Reverse armor naming convention
	var newArmors []image.Texture
	for _, a := range pack.Armors {
		a.Name = bedrockArmorReplacer.Replace(a.Name)
		newArmors = append(newArmors, a)
	}
	newPack.Armors = newArmors

	if err := newPack.WriteZipWithVersion(output, version); err != nil {
		return err
	}
	fmt.Println("zip file written to:", output)
	return nil
}
