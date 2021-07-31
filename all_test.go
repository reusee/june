// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

// +build !step1 !step2

package june

import (
	"github.com/reusee/june/codec"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/file"
	"github.com/reusee/june/filebase"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/naming"
	"github.com/reusee/june/storedisk"
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

	"testing"
)

func Test_codec_TestCodecAESGCM(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestCodecAESGCM)
}

func Test_codec_TestCodecSnappy(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestCodecSnappy)
}

func Test_codec_TestCodecStacked(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestCodecStacked)
}

func Test_codec_TestHybridCompressed(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestHybridCompressed)
}

func Test_codec_TestHybridSnappy(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestHybridSnappy)
}

func Test_codec_TestHybridZstd(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestHybridZstd)
}

func Test_codec_TestOnion(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestOnion)
}

func Test_codec_TestOnion2(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestOnion2)
}

func Test_codec_TestSegmented(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestSegmented)
}

func Test_codec_TestSnappyStream(t *testing.T) {
	t.Parallel()
	runTest(t, codec.TestSnappyStream)
}

func Test_entity_TestDelete(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestDelete)
}

func Test_entity_TestEmbeddingSameObject(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestEmbeddingSameObject)
}

func Test_entity_TestFetch(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestFetch)
}

func Test_entity_TestGC(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestGC)
}

func Test_entity_TestGCWithEmptyIndex(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestGCWithEmptyIndex)
}

func Test_entity_TestIndex(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestIndex)
}

func Test_entity_TestIntegrity(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestIntegrity)
}

func Test_entity_TestLock(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestLock)
}

func Test_entity_TestNewName(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestNewName)
}

func Test_entity_TestPush(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestPush)
}

func Test_entity_TestSave(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestSave)
}

func Test_entity_TestSummary(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestSummary)
}

func Test_entity_TestSummaryUpdate(t *testing.T) {
	t.Parallel()
	runTest(t, entity.TestSummaryUpdate)
}

func Test_file_TestBuild(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestBuild)
}

func Test_file_TestBuildMerge(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestBuildMerge)
}

func Test_file_TestFileFS(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestFileFS)
}

func Test_file_TestFileWithTx(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestFileWithTx)
}

func Test_file_TestIterDiskCancelWaitTree(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestIterDiskCancelWaitTree)
}

func Test_file_TestIterIgnore(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestIterIgnore)
}

func Test_file_TestPack(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestPack)
}

func Test_file_TestPushFile(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestPushFile)
}

func Test_file_TestSave(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestSave)
}

func Test_file_TestSymlink(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestSymlink)
}

func Test_file_TestUpdate(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestUpdate)
}

func Test_file_TestZip(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestZip)
}

func Test_file_TestZipFile(t *testing.T) {
	t.Parallel()
	runTest(t, file.TestZipFile)
}

func Test_filebase_TestContent(t *testing.T) {
	t.Parallel()
	runTest(t, filebase.TestContent)
}

func Test_filebase_TestContentSB(t *testing.T) {
	t.Parallel()
	runTest(t, filebase.TestContentSB)
}

func Test_fsys_TestRestrictedPath(t *testing.T) {
	t.Parallel()
	runTest(t, fsys.TestRestrictedPath)
}

func Test_fsys_TestShuffleDir(t *testing.T) {
	t.Parallel()
	runTest(t, fsys.TestShuffleDir)
}

func Test_fsys_TestWatch(t *testing.T) {
	t.Parallel()
	runTest(t, fsys.TestWatch)
}

func Test_index_TestIdxUnknown(t *testing.T) {
	t.Parallel()
	runTest(t, index.TestIdxUnknown)
}

func Test_index_TestTuple(t *testing.T) {
	t.Parallel()
	runTest(t, index.TestTuple)
}

func Test_key_TestHashFromString(t *testing.T) {
	t.Parallel()
	runTest(t, key.TestHashFromString)
}

func Test_key_TestKeyFromString(t *testing.T) {
	t.Parallel()
	runTest(t, key.TestKeyFromString)
}

func Test_key_TestKeyJSON(t *testing.T) {
	t.Parallel()
	runTest(t, key.TestKeyJSON)
}

func Test_key_TestKeyPath(t *testing.T) {
	t.Parallel()
	runTest(t, key.TestKeyPath)
}

func Test_key_TestNamespaceFromString(t *testing.T) {
	t.Parallel()
	runTest(t, key.TestNamespaceFromString)
}

func Test_naming_TestTypeName(t *testing.T) {
	t.Parallel()
	runTest(t, naming.TestTypeName)
}

func Test_storedisk_TestStore(t *testing.T) {
	t.Parallel()
	runTest(t, storedisk.TestStore)
}

func Test_storedisk_TestStoreSoftDelete(t *testing.T) {
	t.Parallel()
	runTest(t, storedisk.TestStoreSoftDelete)
}

func Test_storemem_TestIndex(t *testing.T) {
	t.Parallel()
	runTest(t, storemem.TestIndex)
}

func Test_storemem_TestStore(t *testing.T) {
	t.Parallel()
	runTest(t, storemem.TestStore)
}

func Test_storenssharded_TestStore(t *testing.T) {
	t.Parallel()
	runTest(t, storenssharded.TestStore)
}

func Test_storeonedrive_TestKV(t *testing.T) {
	t.Parallel()
	runTest(t, storeonedrive.TestKV)
}

func Test_storepebble_TestBatchIndex(t *testing.T) {
	t.Parallel()
	runTest(t, storepebble.TestBatchIndex)
}

func Test_storepebble_TestBatchKV(t *testing.T) {
	t.Parallel()
	runTest(t, storepebble.TestBatchKV)
}

func Test_storepebble_TestIndex(t *testing.T) {
	t.Parallel()
	runTest(t, storepebble.TestIndex)
}

func Test_storepebble_TestKV(t *testing.T) {
	t.Parallel()
	runTest(t, storepebble.TestKV)
}

func Test_storepebble_TestMixedIndex(t *testing.T) {
	t.Parallel()
	runTest(t, storepebble.TestMixedIndex)
}

func Test_storepebble_TestMixedKV(t *testing.T) {
	t.Parallel()
	runTest(t, storepebble.TestMixedKV)
}

func Test_stores3_TestKV(t *testing.T) {
	t.Parallel()
	runTest(t, stores3.TestKV)
}

func Test_storesqlite_TestKV(t *testing.T) {
	t.Parallel()
	runTest(t, storesqlite.TestKV)
}

func Test_storestacked_TestStore(t *testing.T) {
	t.Parallel()
	runTest(t, storestacked.TestStore)
}

func Test_storetap_TestStore(t *testing.T) {
	t.Parallel()
	runTest(t, storetap.TestStore)
}

func Test_sys_TestTesting(t *testing.T) {
	t.Parallel()
	runTest(t, sys.TestTesting)
}

func Test_tx_TestPebbleTx(t *testing.T) {
	t.Parallel()
	runTest(t, tx.TestPebbleTx)
}

func Test_tx_TestPebbleTxEntityDelete(t *testing.T) {
	t.Parallel()
	runTest(t, tx.TestPebbleTxEntityDelete)
}

func Test_vars_TestVars(t *testing.T) {
	t.Parallel()
	runTest(t, vars.TestVars)
}
