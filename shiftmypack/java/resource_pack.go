package java

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/xAstralMCALT/shiftmypack/shiftmypack/image"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/internal/fsutil"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/internal/logger"
)

type ResourcePack struct {
	Name string
	r    *zip.Reader

	Skies []image.Texture

	PackIcon  image.Texture
	Icons     image.Texture
	Particles image.Texture

	Items  []image.Texture
	Blocks []image.Texture
	Armors []image.Texture
}

func NewResourcePack(filePath string) (ResourcePack, error) {
	pck := ResourcePack{}
	pck.Name = filepath.Base(filePath)

	r, err := fsutil.OpenZip(filePath)
	if err != nil {
		return ResourcePack{}, errors.New("error opening java pack zip: " + err.Error())
	}

	path, found := fsutil.FindDirectory(r, "world0")
	if found {
		pck.Skies, err = findSkies(r, path)
		if err != nil {
			logger.Debugf("error trying to find skies: %s", err)
		}
	}

	pck.PackIcon, err = image.NewTextureFS(r, "pack.png", false)
	if err != nil {
		logger.Debugf("error trying to find icon: %s", err)
	}

	texturesPath, found := fsutil.FindDirectory(r, "textures")
	if !found {
		return ResourcePack{}, errors.New("could not find textures path, stopping progress.")
	}

	pck.Icons, err = image.NewTextureFS(r, texturesPath+"/gui/icons.png", true)
	if err != nil {
		logger.Debugf("error trying to find icon: %s", err)
	}
	pck.Particles, err = image.NewTextureFS(r, texturesPath+"/particle/particles.png", false)
	if err != nil {
		logger.Debugf("error trying to find icon: %s", err)
	}

	pck.Items, err = image.DirectoryTexturesFS(r, texturesPath+"/items", true)
	if err != nil {
		logger.Debugf("error trying to find items: %s", err)
	}
	pck.Blocks, err = image.DirectoryTexturesFS(r, texturesPath+"/blocks", false)
	if err != nil {
		logger.Debugf("error trying to find items: %s", err)
	}
	pck.Armors, err = image.DirectoryTexturesFS(r, texturesPath+"/models/armor", false)
	if err != nil {
		logger.Debugf("error trying to find items: %s", err)
	}

	pck.r = r
	return pck, nil
}

func (pck ResourcePack) WriteZip(output string) error {
	return pck.WriteZipWithVersion(output, "1.20")
}

func (pck ResourcePack) WriteZipWithVersion(output string, version string) error {
	f, w, err := fsutil.CreateZip(output)
	if err != nil {
		return errors.New("error creating new zip file: " + err.Error())
	}

	// Write pack.mcmeta
	if err := writeMcmetaWithVersion(w, version); err != nil {
		return err
	}

	pck.PackIcon.Write(w, "pack.png")
	pck.Icons.Write(w, "assets/minecraft/textures/gui/icons.png")
	pck.Particles.Write(w, "assets/minecraft/textures/particle/particles.png")
	writeTextures(w, pck.Skies, "assets/minecraft/textures/environment")
	writeTextures(w, pck.Items, "assets/minecraft/textures/items")
	writeTextures(w, pck.Blocks, "assets/minecraft/textures/blocks")
	writeTextures(w, pck.Armors, "assets/minecraft/textures/models/armor")

	w.Close()
	f.Close()
	return nil
}

func versionToPackFormat(version string) int {
	switch version {
	case "1.12":
		return 2
	case "1.13":
		return 4
	case "1.16":
		return 5
	case "1.17":
		return 6
	case "1.18":
		return 9
	case "1.19":
		return 12
	case "1.20":
		return 15
	default:
		return 15
	}
}

func writeMcmeta(w *zip.Writer) error {
	return writeMcmetaWithVersion(w, "1.20")
}

func writeMcmetaWithVersion(w *zip.Writer, version string) error {
	packFormat := versionToPackFormat(version)
	metadata := map[string]interface{}{
		"pack": map[string]interface{}{
			"pack_format": packFormat,
			"description": "",
		},
	}

	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return errors.New("error marshaling pack.mcmeta: " + err.Error())
	}

	f, err := w.Create("pack.mcmeta")
	if err != nil {
		return errors.New("error creating pack.mcmeta: " + err.Error())
	}

	_, err = f.Write(jsonBytes)
	if err != nil {
		return errors.New("error writing pack.mcmeta: " + err.Error())
	}

	return nil
}

func writeTextures(w *zip.Writer, textures []Texture, path string) {
	for _, t := range textures {
		if t == (Texture{}) {
			continue
		}
		output := filepath.Join(path, t.Name)
		t.Write(w, output)
	}
}

type Texture = image.Texture
