package util

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

// RunInlineJS executes rule-side "@js:" snippets with input bound to variable `r`.
func RunInlineJS(jsCode, input string) (string, error) {
	code := strings.TrimSpace(jsCode)
	if code == "" {
		return input, nil
	}

	vm := goja.New()
	if err := vm.Set("__input", input); err != nil {
		return "", fmt.Errorf("set js input: %w", err)
	}

	script := `function __sonovel_inline(r) {
` + code + `
return r;
}
__sonovel_inline(__input);`

	v, err := vm.RunString(script)
	if err != nil {
		return "", err
	}
	return v.String(), nil
}
