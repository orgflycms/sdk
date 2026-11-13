# sdk — контракт плагина flycms

Открытый репозиторий. Здесь всё, что нужно, чтобы написать плагин для
магазина на flycms: контракт `sdk.go`, пример `example/`, правила ниже.

## Как устроен плагин

Папка с Go-файлами одного пакета и функцией `Init(api sdk.API) error`.
Папка называется полным именем плагина `<пространство>.<имя>`, например
`vitams.promo`: пространство имён партнёр выбирает при регистрации в
кабинете, оно — префикс всех его плагинов. Пакет внутри назван коротким
именем: `package promo`. Ядро магазина читает папку интерпретатором при
старте и по кнопке «Перечитать» в разделе «Плагины» — компилировать и
перезапускать ничего не нужно.

Что даёт `sdk.API` — в [sdk.go](sdk.go), там комментарий к каждому методу.
Живой пример — [example/](example/): пункт меню и страница.

Плагину доступны этот пакет и часть стандартной библиотеки Go — список
ниже. Сторонних модулей нет: их некому скачать.

## Что доступно плагину

Плагин исполняется интерпретатором внутри процесса магазина и получает не
всю стандартную библиотеку, а явный список: пакет `sdk` и 62 пакета ниже, в
каждом — проверенные имена, а не всё подряд. Импорт другого пакета — ошибка
загрузки «пакет … запрещён в песочнице плагина»; имя, которого нет в
списке, — тоже ошибка загрузки. Обе видны в разделе «Плагины». Свои файлы
плагин импортирует только из своей папки.

```text
bufio  bytes  compress/flate  compress/gzip  compress/zlib
container/heap  container/list  context
crypto  crypto/aes  crypto/cipher  crypto/ecdh  crypto/ecdsa  crypto/ed25519
crypto/elliptic  crypto/hmac  crypto/md5  crypto/rand  crypto/rsa  crypto/sha1
crypto/sha256  crypto/sha512  crypto/subtle  crypto/tls  crypto/x509
crypto/x509/pkix
encoding  encoding/base32  encoding/base64  encoding/binary  encoding/csv
encoding/hex  encoding/json  encoding/pem  encoding/xml
errors  fmt  hash  hash/adler32  hash/crc32  hash/crc64  hash/fnv  hash/maphash
html  html/template  io  math  math/big  math/bits  math/rand/v2
mime/quotedprintable  net/http  net/url  path  regexp  sort  strconv  strings
sync  time  unicode/utf16  unicode/utf8
```

Интерпретатор знает таблицу стандартной библиотеки **Go 1.22**: функций,
появившихся в Go 1.23 и позже, у плагина нет. `go build` с `go 1.27` из
`go.mod` их пропустит — проверяйте плагин загрузкой в магазин.

### Чего нет и почему

- **Файлов, процессов, окружения**: `os` и его подпакеты, `os/exec`,
  `syscall`, `io/fs`, `io/ioutil`, `path/filepath`, `embed`, `archive/*`,
  `mime/multipart`, `database/sql` (база — через `api.DB()`). Плагин живёт в
  процессе магазина: через файлы ему достались бы ключи и настройки, через
  процессы — сервер целиком.
- **Сети, кроме клиента HTTP(S)**: `net`, `net/smtp`, `net/mail`,
  `net/rpc`, `net/http/*` (`httptest`, `httputil`, `cookiejar`),
  `tls.Dial` и `tls.Listen`. Своих серверов и сокетов у плагина нет: адреса
  он получает через `api.Handle`.
- **Горутин и отложенных вызовов.** Оператор `go` запрещён во всех файлах
  плагина, даже в функции, которую никто не вызывает: ошибка загрузки
  укажет файл и строку. Нет `time.AfterFunc`, `context.AfterFunc` и
  `sync.WaitGroup`. Почему: магазин ловит панику плагина в `Init`, в
  запросе и в корзине — она стоит плагину ответа или скидки, не магазина.
  Паника в отдельной горутине эту защиту обходит и останавливает весь
  магазин. Каналы, мьютексы, `sync.Once`, `time.Sleep`, `time.After`,
  `NewTimer`, `NewTicker`, `Tick` — можно.
- **Обхода интерпретатора**: `reflect`, `unsafe`, `runtime` и подпакеты,
  `plugin`, `go/*`, `debug/*`, `testing`, `expvar`.
- **Журнала**: `log` и `log/slog` закрыты, `fmt.Print*` пишет в никуда.
  Показать что-то владельцу можно на своей странице в админке.
- **Прочего, чего нет в списке**, даже безобидного: `slices`, `maps`,
  `cmp`, `unicode` (есть `unicode/utf8` и `unicode/utf16`), `text/template`
  (есть `html/template`), `math/rand` (есть `math/rand/v2`). Пакет
  открывается только правкой ядра магазина.
- Внутри разрешённых пакетов убрано немногое: `crypto/rand.Reader` (есть
  `rand.Read`), `crypto.RegisterHash`, `time.LoadLocation` и `time.Local`
  (часовой пояс — `time.FixedZone`, `time.UTC` или
  `time.LoadLocationFromTZData` с данными в памяти), `x509.SystemCertPool`,
  `x509.SetFallbackRoots`, `x509.VerifyOptions`.

### `net/http` — обёртка

Имена те же, набор урезан: только клиент и то, что нужно обработчику.

Есть: `http.Get`, `Head`, `Post`, `PostForm` (таймаут 30 секунд),
`NewRequest`, `NewRequestWithContext`; `http.Client` с полями `Transport`,
`CheckRedirect`, `Jar`, `Timeout` и методами `Do`, `Get`, `Head`, `Post`,
`PostForm`, `CloseIdleConnections`; `http.Transport` с полями
`TLSClientConfig`, `DisableKeepAlives`, `DisableCompression`,
`ForceAttemptHTTP2`, `MaxIdleConns`, `MaxIdleConnsPerHost`,
`MaxConnsPerHost`, `IdleConnTimeout`, `TLSHandshakeTimeout`,
`ExpectContinueTimeout`, `ResponseHeaderTimeout`, `MaxResponseHeaderBytes` и
методами `RoundTrip`, `CloseIdleConnections`; `Header`, `Cookie`,
`SetCookie`, `Error`, `Redirect`, `NotFound`, `MaxBytesReader`,
`StatusText`, `DetectContentType`, константы методов и статусов.

Нет из привычного:

- серверной стороны: `ListenAndServe*`, `Serve*`, `Server`, `ServeMux`,
  `NewServeMux`, `Handle`, `HandleFunc`, `Handler`, `StripPrefix`,
  `TimeoutHandler`, `MaxBytesHandler`, `RedirectHandler`,
  `NotFoundHandler`. Адреса — только `api.Handle` с `http.HandlerFunc`;
- файлов: `FileServer`, `ServeFile`, `ServeContent`, `Dir`, `FS`,
  `NewFileTransport`; файлов из формы: `FormFile`, `ParseMultipartForm`,
  `MultipartReader`, поля `MultipartForm`. `FormValue` и `PostFormValue`
  читают только строку запроса и `application/x-www-form-urlencoded`;
- сырых соединений: `Hijacker`, `ResponseController`, `Pusher`, у
  `Transport` — `Proxy`, `DialContext` и остальных `Dial*`;
- глобальных объектов: `DefaultClient`, `DefaultTransport`,
  `DefaultServeMux`, `ProxyFromEnvironment`; своего `RoundTripper` в
  `Client` не подставить;
- разбора и записи потока: `ReadRequest`, `ReadResponse`, `Request.Write`,
  `Response.Write`.

Поля `Transport` задаются до первого запроса: позже их изменения не
действуют. Тело исходящего запроса читается в память целиком до отправки —
потоковой выгрузки нет. `http.Request` — та же обёртка: поля метода,
адреса, заголовков, тела и форм, `Host`, `RemoteAddr`, `TLS`; методы
`Context`, `WithContext`, `Clone`, `Cookie`, `Cookies`, `AddCookie`,
`UserAgent`, `Referer`, `BasicAuth`, `SetBasicAuth`, `PathValue`,
`SetPathValue`, `ParseForm`, `FormValue`, `PostFormValue`. У ответа — поля
`Status`, `StatusCode`, `Header`, `Body` и прочие, методы `Cookies` и
`Location`.

HTTPS со своим корневым сертификатом: `x509.NewCertPool()` →
`AppendCertsFromPEM` (или `AddCert`) → `&tls.Config{RootCAs: pool}` →
`&http.Transport{TLSClientConfig: cfg}`. Клиентский сертификат —
`tls.X509KeyPair` из PEM в памяти, в `Certificates`. В `tls.Config` только
данные: `RootCAs`, `ClientCAs`, `Certificates`, `ServerName`, `NextProtos`,
`InsecureSkipVerify`, `MinVersion`, `MaxVersion`, `CipherSuites`,
`CurvePreferences`, `SessionTicketsDisabled`, `Renegotiation`; колбэков
(`VerifyPeerCertificate`, `GetClientCertificate` и других), `Rand`, `Time`
и `ClientSessionCache` нет. У пула корней — `AppendCertsFromPEM`,
`AddCert`, `Clone`, `Equal`, `Subjects`, без `AddCertWithConstraint`. Нет
`tls.LoadX509KeyPair` и `NewLRUClientSessionCache`.

Контексты — `context.Background`, `context.TODO`, контекст запроса
(`r.Context()`), контекст перехватчика корзины и то, что из них сделали
`context.With*`. Своя реализация `context.Context` не принимается:
`context.With*`, `NewRequestWithContext`, `WithContext`, `Clone` и
`api.DB()` на ней паникуют. Почему: стандартная библиотека зовёт методы
контекста в своих горутинах, и паника там уронила бы магазин.

### `html/template` — обёртка

Шаблоны — только из строк в памяти: `template.New`, `Must`, у шаблона
`Parse`, `Funcs`, `Delims`, `Option`, `New`, `Lookup`, `Templates`, `Clone`,
`Name`, `DefinedTemplates`, `Execute`, `ExecuteTemplate`; типы `HTML`,
`CSS`, `JS`, `URL` и другие, `FuncMap`, функции экранирования. Нет
`ParseFiles`, `ParseGlob`, `ParseFS` — ни у пакета, ни у шаблона, — а также
`AddParseTree`, поля `Tree` и констант `Err*`. Шаблон держите в коде
строкой: `embed` тоже закрыт.

## Проверить у себя

```sh
go build ./...
```

`go.mod` здесь для этого и для подсказок редактора. `go build` проверяет
только, что код компилируется; запреты выше — закрытые пакеты, оператор
`go`, функции новее Go 1.22 — видны только при загрузке. Положите папку
плагина в папку плагинов магазина — `/opt/flycms/plugins/` у магазина,
поставленного `install.sh`, — и нажмите «Перечитать» в разделе «Плагины»:
ошибка загрузки видна там же.

## Опубликовать в маркетплейсе

Кабинет партнёра на сайте flycms: регистрация с подтверждением почты и
выбором пространства имён, форма заявки, модерация. Подробный разбор с
примерами — раздел «Документация» на сайте. Заявка — короткое имя (префикс
подставится сам), архив `.tar.gz` или `.zip` с файлами плагина, до пяти
картинок и версия. Что проверяется до модерации:

- в архиве только `*.go` (без `_test.go`), `README.md`, `plugin.json`;
  без подпапок; до 1 МБ;
- имя пакета равно короткому имени плагина, есть `func Init(`;
- версия — semver (`1.2.0`) и больше предыдущей опубликованной;
- картинки — jpeg, png или webp до 2 МБ.

`plugin.json` — карточка плагина, как в [example/plugin.json](example/plugin.json):
`name` (полное, с пространством), `title`, `description`, `version`. Имя
плагина после первой публикации не меняется: на нём папки в магазинах.

Модератор читает код и одобряет или возвращает с замечанием — оно видно в
кабинете. Правки опубликованного плагина — та же форма, новая заявка;
карточка показывает последнее одобренное.
