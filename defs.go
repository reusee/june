// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package ling

import (
	"github.com/reusee/dscope"
	"github.com/reusee/ling/v2/clock"
	"github.com/reusee/ling/v2/config"
	"github.com/reusee/ling/v2/entity"
	"github.com/reusee/ling/v2/file"
	"github.com/reusee/ling/v2/filebase"
	"github.com/reusee/ling/v2/flush"
	"github.com/reusee/ling/v2/fsys"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/naming"
	"github.com/reusee/ling/v2/store"
	"github.com/reusee/ling/v2/storedisk"
	"github.com/reusee/ling/v2/storekv"
	"github.com/reusee/ling/v2/storemem"
	"github.com/reusee/ling/v2/storenssharded"
	"github.com/reusee/ling/v2/storeonedrive"
	"github.com/reusee/ling/v2/storepebble"
	"github.com/reusee/ling/v2/stores3"
	"github.com/reusee/ling/v2/storestacked"
	"github.com/reusee/ling/v2/storetap"
	"github.com/reusee/ling/v2/sys"
	"github.com/reusee/ling/v2/tx"
	"github.com/reusee/ling/v2/vars"
	"github.com/reusee/ling/v2/virtualfs"
)

var Defs = dscope.Methods(
	clock.Def{},
	config.Def{},
	entity.Def{},
	file.Def{},
	filebase.Def{},
	flush.Def{},
	fsys.Def{},
	index.Def{},
	key.Def{},
	naming.Def{},
	store.Def{},
	storedisk.Def{},
	storekv.Def{},
	storemem.Def{},
	storenssharded.Def{},
	storeonedrive.Def{},
	storepebble.Def{},
	stores3.Def{},
	storestacked.Def{},
	storetap.Def{},
	sys.Def{},
	tx.Def{},
	vars.Def{},
	virtualfs.Def{},
)
