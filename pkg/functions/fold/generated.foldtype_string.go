// Code generated by "stringer -type=FoldType -output generated.foldtype_string.go"; DO NOT EDIT.

package fold

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UnknownFold-0]
	_ = x[FoldTypeR-1]
	_ = x[FoldTypeL-2]
	_ = x[FoldTypeT-3]
	_ = x[FoldTypeI-4]
}

const _FoldType_name = "UnknownFoldFoldTypeRFoldTypeLFoldTypeTFoldTypeI"

var _FoldType_index = [...]uint8{0, 11, 20, 29, 38, 47}

func (i FoldType) String() string {
	if i < 0 || i >= FoldType(len(_FoldType_index)-1) {
		return "FoldType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FoldType_name[_FoldType_index[i]:_FoldType_index[i+1]]
}
