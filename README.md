# updateric

Ենթադրենք, թե ունենք պայմանական `programma` անունով մի ծրագիր, 
որն աշխատում է հազարավոր օգտագործողների մեքենաների վրա, և մեզ 
անհրաժեշտ է իրականացնել մի գործիք, որը պարբերաբար այն նույն
օգտագործողների մեքենաներում տեղադրում է `programma`֊ի նոր 
թողարկված տարբերակները։

Պարզության համար ընդունենք, որ `programma`֊ն տեղադրված է 
օգտագործողի `$HOME/programma` պանակում, և բաղկացած է նվազագույնը 
երկու ֆայլից․ `version.json` և `programma.py`։ Առաջինում գրածված 
է ծրագրի ընթացիկ տարբերակը։ Օրինակ․

```json
{
    "major": 0,
    "minor": 1
}
```

Երկրորդը կատարվող ֆայլն է։ Մոտավորապես այսպիսին․

```Python
#!/usr/bin/python

if __name__ == '__main__':
    print('I am Programma.')
```

Մեր `programma` ծրագի նոր թողարկումները հրապարակվում են համացանցի 
ինչ֊որ պահոցում։ Թող որ `https://some.web.storge`֊ը այդ պահոցի 
հասցեն է, որտեղ տեղադրվում են երկու ֆայլ․ `release-info.json` և 
`programma.zip`։ Առաջինը տեղեկություն է տալիս թողարկման մասին։ 
Օրինակ․

```json
{
    "version": {
        "major": 0,
        "minor": 2
    },
    "url": "https://some.web.storge/programma.zip",
    "sha1": "fe90cf2b46ac87d68a547602924b35cb11f768e7"
}
```

Իսկ երկրորդը `zip` արխիվ է` հետևյալ կառուցվածքով․

```
programma
├── programma.py
└── version.json
```

Նոր թողարկման առկայությունը ստուգող և օգտագործողի մոտ տեղադրված
հին ծրագիրը նորով փոխարինող գործիքին, բնականաբար, կանվանենք 
`updater`։ Այն աշխատելու է մոտավորապես հետևյալ քայլերով․
1. Ինտերնետից ներբեռնել `release-info.json` ֆայլը ու կարդալ դրա պարունակությունը,
2. Կարդալ տեղադրված ծրագրի `version.json` ֆայլի պարունակությունը,
3. Համեմատել _լոկալ_ ու _թողարկված_ տարբերակները,
4. Եթե նոր թողարկման տարբերակը փոքր կամ հավասար է տեղադրվածինից, ապա ավարտել աշխատանքը։
5. Հակառակ դեպքում ներբեռնել `programma.zip` ֆայլը և դրա պարունակությամբ փոխարինել օգտագործողի մոտ տեղադրվածը։

Ստեղծում եմ `updater` պանակը, իսկ դրա մեջ՝ նույն անունով Գո մոդուլը․

```bash
$ cd Projects
$ mkdir updater
$ cd updater
$ go mod init updater
```

Նույն տեղում ստեղծում եմ `updater.go` ֆայլը՝ հետևյալ սկզբնական պարունակությամբ․

```Go
package main

func main() {
    println("Updater for Programma.")
}
```

Ճիշտ աշխատելու համար `updater`֊ը պետք է իմանա երկու բան․ ա) թե 
օգտագործողի մոտ որտեղ է տեղադրված `programma`֊ն, և բ) թե համացանցում
որտեղ են հրապարակվում `programma`֊ի հերթական թողարկումները։ Այս
երկու պարամետրերի համար նախատեսել եմ `config.json` ֆայլը։ Օրինակ, 
այսպես․

```json
{
    "application-path": "/this/is/application/path",
    "release-info-url": "https:///this/is/release/info.url"
}
```

Առաջին հերթին `updater`֊ը պետք է կարդա այս ֆայլը ու իր մոտ պահի 
դրանում նշված տվյալները։ Սահմանեմ `configuration` ստրուկտուրան․

```Go
type configuration struct {
	ApplicationPath string `json:"application-path"`
	UpdateInfoUrl   string `json:"update-info-url"`
}
```

Ֆայլը կարդալու և այս ստրուկտուրայի նմուշն արժեքավորելու համար 
սահմանեմ `readJsonFile` ֆունկցիան, որն իր առաջին արգումենտով 
տրված JSON ֆայլի պարունակությունը կարդում է և արժեքավորում է 
երկրորդ արգումենտով տրված օբյեկտը։ Ահա այն․

```Go
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
```

Այս նույն ֆունկցիան եմ օգտագործելու նաև `release-info.json` և 
`version.json` ֆայլերի պարունակությունները կարդալու և համապատասխան 
օբյեկտներն արժեքավորելու համար։

Ինտերնետից `release-info.json` ֆայլը ներբեռնելու ու դրա 
պարունակությունը կարդալու համար օգտագործելու եմ 
[grab](https://github.com/cavaliergopher/grab) գրադարանը։ 
Ավելացնեմ այն իմ պրոյեկտին․

```bash
$ go get github.com/cavaliergopher/grab/v3
```

Հետո սահմանեմ `downloadFile` ֆունկցիան, որը տրված URL֊ից ֆայլը
ներբեռնում է ժամանակավոր ֆայլերի պանակում և ապա վերադարձնում է 
ներբեռնված ֆայլի ճանապարհը։

```Go
func downloadFile(from string) (string, error) {
	resp, err := grab.Get(os.TempDir(), from)
	if err != nil {
		return "", err
	}
	return resp.Filename, nil
}
```

`release-info.json` ֆայլի պարունակությունը ներկայացնելու համար 
սահմանում եմ `release` ստրուկտուրան․

```Go
type release struct {
	Version version `json:"version"`
	Url  string `json:"url"`
    Sha1 string `json:"sha1"`
}
```

Այստեղ օգտագործված `version` տիպը նույնպես ստրուկտուրա է․

```Go
type version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}
```
