package main

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/cavaliergopher/grab/v3"
)

// updaer֊ին անհրաժեշտ երկու պարամետրերը
type configuration struct {
	// տեղադրված ծրագրի պանակը
	ApplicationPath string `json:"application-path"`
	// նոր թողարկման հասցեն
	ReleaseInfoUrl string `json:"release-info-url"`
}

// ծրագրի տարբերակի ստրուկտուրան
type version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

// isNewer ֆունկցիան պարզում է, թե արդյո՞ք v֊ն մեծ է x֊ից
func (v version) isNewer(x version) bool {
	if v.Major > x.Major {
		return true
	}

	if v.Major < x.Major {
		return false
	}

	return v.Minor > x.Minor
}

// թողարկման մասին տեղեկություննները
type manifest struct {
	// թողարկման տարբերակը
	Version version `json:"version"`
	// ծրագրի արխիվի հասցեն է ․․․
	Url string `json:"url"`
	// ․․․ և sha1 կոդը
	Sha1 string `json:"sha1"`
}

func main() {
	var conf configuration
	if readJsonFile("./config.json", &conf) != nil {
		log.Fatal("Չկարողացա կարդալ պարամետրերի ֆայլը։")
	}

	relInfo, err := downloadFile(conf.ReleaseInfoUrl)
	if err != nil {
		log.Fatal("Չկարողացա ներբեռնել release-info.json ֆայլը։")
	}

	var mf manifest
	if readJsonFile(relInfo, &mf) != nil {
		log.Fatal("Չկարողացա կարդալ release-info.json ֆայլը։")
	}

	var vo version
	verFile := filepath.Join(conf.ApplicationPath, "version.json")
	if readJsonFile(verFile, &vo) != nil {
		log.Fatal("Չկարողացա կարդալ version.json ֆայլը։")
	}

	if mf.Version.isNewer(vo) {
		bundle, err := downloadFile(mf.Url)
		if err != nil {
			log.Fatalf("Չկարողացա ներբեռնել %s ֆայլը։", mf.Url)
		}

		shab, err := calculateSha1(bundle)
		if err != nil {
			log.Fatalf("Չկարողացա հաշվել %s֊ի SHA1֊ը", bundle)
		}

		if shab != mf.Sha1 {
			log.Fatal("Ներբեռնված արխիվի SHA1֊ը չի համընկնում հայտարարվածին։")
		}

		backupPath := conf.ApplicationPath + "_backup"
		err = os.Rename(conf.ApplicationPath, backupPath)
		if err != nil {
			log.Fatal("Չկարողացա անվանափոխել տեղադրված ծրագրի պանակը։")
		}

		err = extractZip(bundle, filepath.Dir(conf.ApplicationPath))
		if err != nil {
			log.Fatal("Չկարողացա բացել ներբեռնված արխիվը։")
			os.Rename(backupPath, conf.ApplicationPath)
		}

		os.RemoveAll(backupPath)
	}

	println("Programma updated")
}

// readJsonFile ֆունկցիան կարդում է JSON ֆայլը և
// արժեքավորում է obj օբյեկտը
func readJsonFile(file string, obj any) error {
	fs, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fs.Close()

	data, err := ioutil.ReadAll(fs)
	if err != nil {
		return err
	}

	if json.Unmarshal(data, obj) != nil {
		return err
	}

	return nil
}

// downloadFile ֆունկցիան ներբեռնում է from հասցեով ֆայլը
// և այն պահում է to պանակում
func downloadFile(from string) (string, error) {
	resp, err := grab.Get(os.TempDir(), from)
	if err != nil {
		return "", err
	}
	return resp.Filename, nil
}

// extractZip ֆունկցիան բացում է file արխիվը և գրում է to պանակում
func extractZip(file, to string) error {
	reader, err := zip.OpenReader(file)
	if err != nil {
		return err
	}

	for _, f := range reader.Reader.File {
		zf, err := f.Open()
		if err != nil {
			return err
		}
		defer zf.Close()

		t := filepath.Join(to, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(t, f.Mode())
		} else {
			const opFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			of, err := os.OpenFile(t, opFlags, f.Mode())
			if err != nil {
				return err
			}
			defer of.Close()

			_, err = io.Copy(of, zf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// calculateSha1 ֆունկցիան հաշվարկում և տողի տեսքով
// վերադարձնում է տրված ֆայլի SHA1-ը
func calculateSha1(ph string) (string, error) {
	f, e := os.Open(ph)
	if e != nil {
		return "", e
	}
	defer f.Close()

	sum := sha1.New()
	_, e = io.Copy(sum, f)
	if e != nil {
		return "", e
	}

	return string(sum.Sum(nil)), nil
}
