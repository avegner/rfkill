// +build mips mips64 ppc64 s390x

//nolint:gochecknoglobals
package rfkill

import "encoding/binary"

var endianess = binary.BigEndian
