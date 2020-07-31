//go:generate ../../.build-tools/esc -o definitions.go -modtime 1489449600 -pkg muc -ignore "Makefile" definitions/

package muc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

type builder struct {
	gtki.Builder
}

func newBuilder(filename string) *builder {
	return newBuilderFromString(filename)
}

func newBuilderFromString(uiName string) *builder {
	return &builder{builderForDefinition(uiName)}
}

func (b *builder) getObj(name string) glibi.Object {
	obj, _ := b.GetObject(name)
	return obj
}

func (b *builder) bindObjects(view interface{}) error {
	if reflect.TypeOf(view).Kind() != reflect.Ptr {
		return errors.New("view must be a pointer")
	}

	elem := reflect.ValueOf(view).Elem()

	dstType := elem.Type()
	if dstType.Kind() != reflect.Struct {
		return errors.New("view must be a pointer to a struct value")
	}

	for i := 0; i < dstType.NumField(); i++ {
		objectID := dstType.Field(i).Tag.Get("gtk-widget")
		if objectID == "" {
			continue
		}

		dstValue := elem.Field(i)

		if !dstValue.CanSet() {
			//Unexported field. This is fine by unsafe pkg documentation
			/* #nosec G103 */
			dstValue = reflect.NewAt(dstValue.Type(), unsafe.Pointer(dstValue.UnsafeAddr())).Elem()
		}

		if !dstValue.CanSet() {
			return errors.New("cant set value")
		}

		object := b.get(objectID)
		v := reflect.ValueOf(object)
		//if dstValue.Type() != v.Type() {
		//	return errors.New("types do not match")
		//}

		dstValue.Set(v)
	}

	return nil
}

func (b *builder) getItem(name string, target interface{}) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		panic("builder.getItem() target argument must be a pointer")
	}
	elem := v.Elem()
	elem.Set(reflect.ValueOf(b.get(name)))
}

//TODO: Why not a map[string]interface{}?
func (b *builder) getItems(args ...interface{}) {
	for len(args) >= 2 {
		name, ok := args[0].(string)
		if !ok {
			panic("string argument expected in builder.getItems()")
		}
		b.getItem(name, args[1])
		args = args[2:]
	}
}

func (b *builder) get(name string) glibi.Object {
	obj, err := b.GetObject(name)
	if err != nil {
		panic("builder.GetObject() failed: " + err.Error())
	}
	return obj
}

const (
	xmlExtension = ".xml"
)

func getDefinitionWithFileFallback(uiName string) string {
	fname := path.Join("/definitions", uiName+xmlExtension)

	embeddedFile, err := FSString(false, fname)
	if err != nil {
		//Enforce the file is embedded, but dont use it.
		panic(fmt.Sprintf("No definition found for %s", uiName))
	}

	fileName := filepath.Join(getActualDefsFolder(), uiName+xmlExtension)
	if fileNotFound(fileName) {
		return embeddedFile
	}

	log.WithField("file", fileName).Debug("Loading definition from local file")
	return readFile(fileName)
}

func builderForDefinition(uiName string) gtki.Builder {
	template := getDefinitionWithFileFallback(uiName)

	builder, err := g.gtk.BuilderNew()
	if err != nil {
		//We cant recover from this
		panic(err)
	}

	//We dont use NewFromString because it doesnt give us an error message
	err = builder.AddFromString(template)
	if err != nil {
		//This is a programming error
		panic(fmt.Sprintf("gui: failed load %s: %s\n", uiName, err.Error()))
	}

	return builder
}

func fileNotFound(fileName string) bool {
	_, fnf := os.Stat(fileName)
	return os.IsNotExist(fnf)
}

func readFile(fileName string) string {
	data, _ := ioutil.ReadFile(filepath.Clean(fileName))
	return string(data)
}

func getActualDefsFolder() string {
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "/gui/muc") {
		return "definitions"
	}
	return "gui/muc/definitions"
}