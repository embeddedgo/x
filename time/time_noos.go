// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import _ "unsafe"

// Set sets the current time.
func Set(t Time) { runtime_setwalltime(t.sec(), t.nsec()) }

//go:linkname runtime_setwalltime runtime.setwalltime
func runtime_setwalltime(sec int64, nsec int32)
