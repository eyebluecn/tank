package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestHello(t *testing.T) {

	split := strings.Split("good", "/")
	fmt.Printf("%v", split)

}
