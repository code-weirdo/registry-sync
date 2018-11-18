package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/heroku/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
)

func main() {
	sourceRegistryUrl := flag.String("source", "", "The URL of the source Docker registry (mandatory)")
	sourceUsername := flag.String("source-user", "", "The username for the source Docker registry")
	sourcePassword := flag.String("source-pass", "", "The password for the source Docker registry")
	destinationRegistryUrl := flag.String("destination", "", "The URL of the destination Docker registry (mandatory)")
	destinationUsername := flag.String("destination-user", "", "The username for the destination Docker registry")
	destinationPassword := flag.String("destination-pass", "", "The password for the destination Docker registry")
	image := flag.String("image", "", "The name of the image to sync from the source to the destination")
	tag := flag.String("tag", "latest", "The tag that should be synced")
	flag.Parse()

	if *sourceRegistryUrl == "" || *destinationRegistryUrl == "" || *image == "" || *tag == "" {
		fmt.Println("For help:")
		fmt.Println("  registry-sync -h")
		os.Exit(0)
	}

	destinationRegistry := GetRegistry(*destinationRegistryUrl, *destinationUsername, *destinationPassword)
	if HasManifest(destinationRegistry, *image, *tag) {
		os.Exit(0)
	}

	sourceRegistry := GetRegistry(*sourceRegistryUrl, *sourceUsername, *sourcePassword)
	sourceManifest := GetManifest(sourceRegistry, *image, *tag)
	layers := GetLayers(sourceManifest)
	for _, layer := range layers {
		if !HasBlob(destinationRegistry, *image, layer) {
			layerContent := GetBlob(sourceRegistry, *image, layer)
			PutBlob(destinationRegistry, *image, layer, layerContent)
		}
	}
	PutManifest(destinationRegistry, *image, *tag, sourceManifest)
}

func GetRegistry(registryUrl, username, password string) *registry.Registry {
	registry, err := registry.New(registryUrl, username, password)
	if err != nil {
		log.Fatal(err)
	}
	return registry
}

func HasManifest(registry *registry.Registry, image, tag string) bool {
	_, err := registry.Manifest(image, tag)
	return err == nil
}

func GetManifest(registry *registry.Registry, image, tag string) *schema1.SignedManifest {
	manifest, err := registry.Manifest(image, tag)
	if err != nil {
		log.Fatal(err)
	}
	return manifest
}

func GetLayers(manifest *schema1.SignedManifest) []digest.Digest {
	var layers []digest.Digest
	for _, layer := range manifest.Manifest.FSLayers {
		layers = append(layers, layer.BlobSum)
	}
	return layers
}

func HasBlob(registry *registry.Registry, image string, blobDigest digest.Digest) bool {
	hasBlob, _ := registry.HasBlob(image, blobDigest)
	return hasBlob
}

func GetBlob(registry *registry.Registry, image string, blobDigest digest.Digest) []byte {
	layer, _ := registry.DownloadBlob(image, blobDigest)
	if layer != nil {
		defer layer.Close()
	}
	bytes, _ := ioutil.ReadAll(layer)
	return bytes
}

func PutBlob(registry *registry.Registry, image string, blobDigest digest.Digest, layerContent []byte) {
	reader := bytes.NewReader(layerContent)
	registry.UploadBlob(image, blobDigest, reader)
}

func PutManifest(registry *registry.Registry, image string, tag string, sourceManifest *schema1.SignedManifest) {
	err := registry.PutManifest(image, tag, sourceManifest)
	if err != nil {
		log.Fatal(err)
	}
}
