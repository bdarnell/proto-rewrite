package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"unsafe"

	_ "github.com/cockroachdb/c-protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// #cgo CXXFLAGS: -std=c++11
// #cgo CPPFLAGS: -I ../../cockroachdb/c-protobuf/internal/src
// #cgo darwin LDFLAGS: -Wl,-undefined -Wl,dynamic_lookup
// #cgo !darwin LDFLAGS: -Wl,-unresolved-symbols=ignore-all
// #include <stdlib.h>
// #include "descriptors.h"
import "C"

var (
	outDir  = flag.String("out", "", "output directory (empty for stdout)")
	filters = flag.String("filters", "",
		"comma-separated list of filters to run")
	protoPath = flag.String("proto_path", "",
		"proto file search path (colon-separated)")
	protoc = flag.String("protoc", "protoc", "path to protoc executable")
)

func loadDescriptors(filenames []string) (*descriptor.FileDescriptorSet, error) {
	tempfile, err := ioutil.TempFile("", "proto-rewrite")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %s", err)
	}
	defer os.Remove(tempfile.Name())

	args := []string{"--descriptor_set_out=" + tempfile.Name()}
	args = append(args, "--proto_path="+*protoPath)
	args = append(args, "--include_imports", "--include_source_info")
	args = append(args, filenames...)
	cmd := exec.Command(*protoc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to run protoc: %s", err)
	}

	descData, err := ioutil.ReadAll(tempfile)
	if err != nil {
		return nil, fmt.Errorf("error reading descriptor: %s", err)
	}

	descriptor := descriptor.FileDescriptorSet{}
	err = proto.Unmarshal(descData, &descriptor)
	if err != nil {
		return nil, fmt.Errorf("error parsing descriptor: %s", err)
	}

	return &descriptor, nil
}

func stripGogoOptions(descriptorSet *descriptor.FileDescriptorSet) {
	for _, fd := range descriptorSet.File {
		if fd.GetPackage() != "proto" {
			continue
		}
		toDelete := -1
		for i, dep := range fd.Dependency {
			if dep == "github.com/gogo/protobuf/gogoproto/gogo.proto" {
				toDelete = i
				break
			}
		}
		if toDelete != -1 {
			fd.Dependency[toDelete] = fd.Dependency[len(fd.Dependency)-1]
			fd.Dependency = fd.Dependency[:len(fd.Dependency)-1]
		}
	}
}

func main() {
	flag.Parse()

	if len(*protoPath) == 0 {
		log.Fatalf("--proto_path is required")
	}

	descriptorSet, err := loadDescriptors(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	stripGogoOptions(descriptorSet)

	reencoded, err := proto.Marshal(descriptorSet)
	if err != nil {
		log.Fatalf("error encoding descriptor set: %s", err)
	}

	cReencoded := C.CString(string(reencoded))
	for _, filename := range flag.Args() {
		cFilename := C.CString(filename)
		cOutput := C.decompile_proto(cReencoded, C.size_t(len(reencoded)),
			cFilename)
		C.free(unsafe.Pointer(cFilename))
		var output string
		if cOutput != nil {
			output = C.GoString(cOutput)
			C.free(unsafe.Pointer(cOutput))
		}
		fmt.Printf("%s", output)
	}
	C.free(unsafe.Pointer(cReencoded))
}
