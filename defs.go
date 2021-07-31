// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package june

import (
	"github.com/reusee/dscope"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/file"
	"github.com/reusee/june/filebase"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/naming"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storedisk"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/june/storenssharded"
	"github.com/reusee/june/storeonedrive"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/june/stores3"
	"github.com/reusee/june/storesqlite"
	"github.com/reusee/june/storestacked"
	"github.com/reusee/june/storetap"
	"github.com/reusee/june/sys"
	"github.com/reusee/june/tx"
	"github.com/reusee/june/vars"
	"github.com/reusee/june/virtualfs"
)

var Defs = dscope.Methods(
	entity.Def{},
	file.Def{},
	filebase.Def{},
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
	storesqlite.Def{},
	storestacked.Def{},
	storetap.Def{},
	sys.Def{},
	tx.Def{},
	vars.Def{},
	virtualfs.Def{},
)
