package assets

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testdata = `\
e12a5c9b0507e23e4c121ac7092a69cf57a74af555c7dc8aa0e8fc5bacb7e3d2  ./assets/FormLabel-8cb728e5.js
ab2c2f615d83eeb63d1d03c7781e04bb15be38167f426de4997da975aca995b4  ./assets/video-757ef459.js
47495bcbc137556fc38bee775b7eb25c0c1ce8726bb9f7b2b35daa0487e49778  ./assets/user-58a25705.js
d9c87b209ae3c3f44d940d6e82cb950569505e765c2db8a23ad170ff4d179694  ./index.html
`

func TestLoad(t *testing.T) {
	assets, err := load(testdata)
	assert.NoError(t, err)
	assert.Len(t, assets, 4)

	assert.Equal(t, "/assets/FormLabel-8cb728e5.js", assets[0].Path)
	assert.Equal(t, "/assets/video-757ef459.js", assets[1].Path)
	assert.Equal(t, "/assets/user-58a25705.js", assets[2].Path)
	assert.Equal(t, "/index.html", assets[3].Path)

	assert.Equal(t, "e12a5c9b0507e23e4c121ac7092a69cf57a74af555c7dc8aa0e8fc5bacb7e3d2", hex.EncodeToString(assets[0].Hash[:]))
	assert.Equal(t, "ab2c2f615d83eeb63d1d03c7781e04bb15be38167f426de4997da975aca995b4", hex.EncodeToString(assets[1].Hash[:]))
	assert.Equal(t, "47495bcbc137556fc38bee775b7eb25c0c1ce8726bb9f7b2b35daa0487e49778", hex.EncodeToString(assets[2].Hash[:]))
	assert.Equal(t, "d9c87b209ae3c3f44d940d6e82cb950569505e765c2db8a23ad170ff4d179694", hex.EncodeToString(assets[3].Hash[:]))
}
