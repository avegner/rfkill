// +build !mips,!mips64,!ppc64,!s390x

package rfkill

import "encoding/binary"

var endianess = binary.LittleEndian
