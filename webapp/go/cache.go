package main

import (
	"github.com/puzpuzpuz/xsync/v3"
)

var iconImageHashCache = xsync.NewMapOf[int64, string]()

var iconImageCache = xsync.NewMapOf[int64, []byte]()
