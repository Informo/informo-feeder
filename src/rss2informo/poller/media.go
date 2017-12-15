package poller

import (
	"mime"
	"regexp"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	"github.com/sirupsen/logrus"
)

var imgRegexp = `<img[^>]+["|']((//|http(s?)://)[^"'>]+)["|']`

func (p *Poller) replaceMedias(content *string) error {
	urls := p.getMediaLinks(*content)
	return p.replaceWithMatrixLink(content, urls)
}

func (p *Poller) replaceWithMatrixLink(content *string, urls []string) (err error) {
	if len(urls) > 0 {
		// map[originalURL]matrixURL
		replacements := make(map[string]string)

		var resp *gomatrix.RespMediaUpload
		for _, url := range urls {
			if resp, err = p.mxClient.UploadLink(url); err != nil {
				return
			}

			logrus.WithFields(logrus.Fields{
				"originalURL": url,
				"mxURL":       resp.ContentURI,
			}).Debug("Replacing media link in content")

			replacements[url] = resp.ContentURI

			time.Sleep(200 * time.Millisecond)
		}

		for origURL, mxURL := range replacements {
			*content = strings.Replace(*content, origURL, mxURL, -1)
		}
	}

	return
}

func (p *Poller) getMediaLinks(content string) (medias []string) {
	medias = make([]string, 0)

	r := regexp.MustCompile(imgRegexp)
	urls := r.FindAllStringSubmatch(content, -1)

	for _, url := range urls {
		if isImage(url[1]) {
			medias = append(medias, url[1])
		}
	}

	return
}

func isImage(fileName string) bool {
	splitName := strings.Split(fileName, ".")
	ext := "." + splitName[len(splitName)-1]
	mimetype := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimetype, "image")
}
