package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StorageTestSuite struct {
	suite.Suite
	tmpDir   string
	storeDir string
}

func (suite *StorageTestSuite) SetupTest() {
	tmpDir, err := ioutil.TempDir("", "cruisemic.storage")
	if err != nil {
		panic(err)
	}
	suite.tmpDir = tmpDir
	suite.storeDir = filepath.Join(tmpDir, "dir")
}

func (suite *StorageTestSuite) TeardownTest() {
	os.RemoveAll(suite.tmpDir)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) TestDirCreation() {
	store, err := NewDiskStorage(suite.storeDir, "", "", nil, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.DirExists(suite.T(), suite.storeDir, "storage directory should exist")
}

func (suite *StorageTestSuite) TestFeedCreation() {
	store, err := NewDiskStorage(suite.storeDir, "", "", nil, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.DirExists(suite.T(), suite.storeDir, "storage directory should exist")
}

func (suite *StorageTestSuite) TestHeader() {
	feedHeaders := map[string]string{
		"empty":    "",
		"header":   "header\ntext",
		"headerLF": "header\ntext\n",
	}

	store, err := NewDiskStorage(suite.storeDir, "test-", ".tab", feedHeaders, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(filepath.Join(suite.storeDir, "test-empty.tab"))
	assert.Nil(suite.T(), err)
	if err == nil {
		assert.Equal(suite.T(), feedHeaders["empty"], string(b), "empty header should not contain header text")
	}
	b, err = ioutil.ReadFile(filepath.Join(suite.storeDir, "test-header.tab"))
	assert.Nil(suite.T(), err)
	if err == nil {
		assert.Equal(suite.T(), feedHeaders["header"]+"\n", string(b), "header text should have LF added")
	}
	b, err = ioutil.ReadFile(filepath.Join(suite.storeDir, "test-headerLF.tab"))
	assert.Nil(suite.T(), err)
	if err == nil {
		assert.Equal(suite.T(), feedHeaders["headerLF"], string(b), "header text should not have LF added")
	}

	// Make sure headers don't get rewritten when files reopened.
	store, err = NewDiskStorage(suite.storeDir, "test-", ".tab", feedHeaders, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}

	b, err = ioutil.ReadFile(filepath.Join(suite.storeDir, "test-empty.tab"))
	assert.Nil(suite.T(), err)
	if err == nil {
		assert.Equal(suite.T(), feedHeaders["empty"], string(b), "reopened empty header should not contain header text")
	}
	b, err = ioutil.ReadFile(filepath.Join(suite.storeDir, "test-header.tab"))
	assert.Nil(suite.T(), err)
	if err == nil {
		assert.Equal(suite.T(), feedHeaders["header"]+"\n", string(b), "reopened header text should have LF added")
	}
	b, err = ioutil.ReadFile(filepath.Join(suite.storeDir, "test-headerLF.tab"))
	assert.Nil(suite.T(), err)
	if err == nil {
		assert.Equal(suite.T(), feedHeaders["headerLF"], string(b), "reopened header text should not have LF added")
	}
}

func (suite *StorageTestSuite) TestWriteStringWithHeader() {
	feedHeaders := map[string]string{"feed": "header\ntext"}
	store, err := NewDiskStorage(suite.storeDir, "test-", ".tab", feedHeaders, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.WriteString("feed", "some text to write\n")
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(filepath.Join(suite.storeDir, "test-feed.tab"))
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.Equal(suite.T(), "header\ntext\nsome text to write\n", string(b), "store.WriteString should write feed text to existing file")
}

func (suite *StorageTestSuite) TestWriteString() {
	store, err := NewDiskStorage(suite.storeDir, "test-", ".tab", nil, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.WriteString("feed", "some text to write\n")
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(filepath.Join(suite.storeDir, "test-feed.tab"))
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.Equal(suite.T(), "some text to write\n", string(b), "store.WriteString should write feed text to new feed file")
}

func (suite *StorageTestSuite) TestWriteStringTwice() {
	store, err := NewDiskStorage(suite.storeDir, "test-", ".tab", nil, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.WriteString("feed", "1\n")
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.WriteString("feed", "2\n")
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.Close()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(filepath.Join(suite.storeDir, "test-feed.tab"))
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.Equal(suite.T(), "1\n2\n", string(b), "store.WriteString should write feed text twice")
}

func (suite *StorageTestSuite) TestFlush() {
	store, err := NewDiskStorage(suite.storeDir, "test-", ".tab", nil, 0)
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	err = store.WriteString("feed", "some text to write\n")
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(filepath.Join(suite.storeDir, "test-feed.tab"))
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.Equal(suite.T(), "", string(b), "buffered text should not be written before calling Flush")
	err = store.Flush()
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	b, err = ioutil.ReadFile(filepath.Join(suite.storeDir, "test-feed.tab"))
	assert.Nil(suite.T(), err)
	if err != nil {
		return
	}
	assert.Equal(suite.T(), "some text to write\n", string(b), "buffered text should be written after calling Flush")
}
