// Code generated by "stringer -type=HookType -output generated.hooktype_string.go"; DO NOT EDIT.

package executor

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UnknownHook-0]
	_ = x[BeforeHook-1]
	_ = x[AfterHook-2]
	_ = x[RunningHook-3]
}

const _HookType_name = "UnknownHookBeforeHookAfterHookRunningHook"

var _HookType_index = [...]uint8{0, 11, 21, 30, 41}

func (i HookType) String() string {
	if i < 0 || i >= HookType(len(_HookType_index)-1) {
		return "HookType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _HookType_name[_HookType_index[i]:_HookType_index[i+1]]
}