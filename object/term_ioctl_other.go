//go:build aix || linux || solaris || zos

package object

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TCGETS
const ioctlWriteTermios = unix.TCSETS
