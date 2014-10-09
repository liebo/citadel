package citadel

import "testing"

func TestParseImageNamePublic(t *testing.T) {
	image := "citadel/foo:latest"
	imageInfo := parseImageName(image)
	if imageInfo.Name != "citadel/foo" {
		t.Fatalf("expected name citadel; received %s", imageInfo.Name)
	}

	if imageInfo.Tag != "latest" {
		t.Fatalf("expected tag latest; received %s", imageInfo.Tag)
	}
}

func TestParseImageNameCustomRegistry(t *testing.T) {
	image := "registry.citadel.com/foo:latest"
	imageInfo := parseImageName(image)
	if imageInfo.Name != "registry.citadel.com/foo" {
		t.Fatalf("expected name registry.citadel.com; received %s", imageInfo.Name)
	}

	if imageInfo.Tag != "latest" {
		t.Fatalf("expected tag latest; received %s", imageInfo.Tag)
	}
}

func TestParseImageNameCustomRegistryPort(t *testing.T) {
	image := "registry.citadel.com:49153/foo:latest"
	imageInfo := parseImageName(image)
	if imageInfo.Name != "registry.citadel.com:49153/foo" {
		t.Fatalf("expected name registry.citadel.com:49153/foo; received %s", imageInfo.Name)
	}

	if imageInfo.Tag != "latest" {
		t.Fatalf("expected tag latest; received %s", imageInfo.Tag)
	}
}
