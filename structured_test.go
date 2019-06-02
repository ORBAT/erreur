package erreur

import (
	"fmt"

	"go.uber.org/zap"
)

func ExampleNew_zap() {
	// How to log errors using zap

	connErr := New("connection error", zap.Int("code", 1234), zap.String("addr", "example.com"))
	// [...] elsewhere in your code
	zap.NewExample().Error("failed to load data", Field(connErr))
	// Output: {"level":"error","msg":"failed to load data","error":{"msg":"connection error","code":1234,"addr":"example.com"}}
}

func ExampleNew_JSON() {
	connErr := New("connection error", zap.Int("code", 1234), zap.String("addr", "example.com"))

	// [...] elsewhere in your code
	structured, _ := AsStructured(connErr)
	fmt.Println(structured.JSON())

	// Output: {"msg":"connection error","code":1234,"addr":"example.com"}
}

func ExampleWrap() {
	const fileName = "someFile"
	someError := String("insufficient permissions")
	err := Wrap(someError, "writing to file failed", zap.String("fileName", fileName))
	zap.NewExample().Error("failed to flush db", Field(err))

	// Output: {"level":"error","msg":"failed to flush db","error":{"msg":"writing to file failed","fileName":"someFile"}}
}

func ExampleWrap_cause() {
	const fileName = "someFile"
	someError := String("insufficient permissions")
	err := Wrap(someError, "writing to file failed", zap.String("fileName", fileName))

	finalErr := Wrap(err, "failed to flush db", zap.String("fieldThatGoes", "ping"))

	zap.NewExample().Error("failed to flush db", Field(finalErr))

	// Output: {"level":"error","msg":"failed to flush db","error":{"msg":"failed to flush db","fieldThatGoes":"ping","cause":{"msg":"writing to file failed","fileName":"someFile"}}}
}