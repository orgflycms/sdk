// Плагин-пример: пункт меню и страница в админке. Это минимум, с которого
// начинается свой плагин; работа с базой и корзиной — в комментариях к
// sdk.API.
package example

import (
	"html/template"
	"net/http"

	"github.com/sitecms/sdk"
)

func Init(api sdk.API) error {
	api.Menu("Пример", "/")
	api.Handle("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		body := `<div class="card p-5 sm:p-6">Плагин «` + template.HTMLEscapeString(api.Name()) + `» работает.</div>`
		api.Page(w, r, "Пример", template.HTML(body))
	})
	return nil
}
