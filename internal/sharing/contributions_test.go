package sharing

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestContribution_ShouldAddUpload(t *testing.T) {
	c := Contribution{
		uploaded: 50,
	}
	c = c.addUpload(350)

	assert.Equal(t, int64(400), c.uploaded)
}

func TestContribution_ShouldAddDownload(t *testing.T) {
	c := Contribution{
		downloaded: 50,
		left:       math.MaxInt64,
	}
	c = c.addDownload(350)

	assert.Equal(t, int64(400), c.downloaded)
}

func TestContribution_ShouldDecreaseLeftOnAddDownload(t *testing.T) {
	c := Contribution{
		downloaded: 0,
		left:       1000,
	}
	c = c.addDownload(300)

	assert.Equal(t, int64(700), c.left)
}

func TestContribution_ShouldNotAddMoreDownloadThanLeft(t *testing.T) {
	c := Contribution{
		downloaded: 0,
		left:       1000,
	}
	c = c.addDownload(5000)

	assert.Equal(t, int64(0), c.left)
	assert.Equal(t, int64(1000), c.downloaded)
}

func TestContribution_ShouldNotAddDownloadIfLeftAlreadyZero(t *testing.T) {
	c := Contribution{
		downloaded: 0,
		left:       0,
	}
	c.addDownload(50)

	assert.Equal(t, int64(0), c.downloaded)
}

func TestContribution_ShouldBeFullyDownloaded(t *testing.T) {
	assert.True(t, Contribution{left: 0}.isDownloadComplete())
	assert.False(t, Contribution{left: 500}.isDownloadComplete())
}

func TestContributions_ShouldAddUpload(t *testing.T) {
	c := Contributions{
		overall: Contribution{},
		session: Contribution{},
	}
	c = c.addUpload(500)

	assert.Equal(t, int64(500), c.overall.uploaded)
	assert.Equal(t, int64(500), c.session.uploaded)
}

func TestContributions_ShouldAddDownload(t *testing.T) {
	c := Contributions{
		overall: Contribution{left: 5000},
		session: Contribution{left: 5000},
	}
	c = c.addDownload(500)

	assert.Equal(t, int64(500), c.overall.downloaded)
	assert.Equal(t, int64(500), c.session.downloaded)
}

func TestContributions_ShouldBeFullyDownloadedBasedOnSession(t *testing.T) {
	complete := Contributions{
		overall: Contribution{left: 5000},
		session: Contribution{},
	}

	assert.True(t, complete.isDownloadComplete())

	incomplete := Contributions{
		overall: Contribution{},
		session: Contribution{left: 5000},
	}

	assert.False(t, incomplete.isDownloadComplete())
}
