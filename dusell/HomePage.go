package dusell

import "../_obj/web"

type homePage struct {
        web.TemplateName
        web.StandardFields
}

var home *homePage = &homePage{
        web.TemplateName(TemplateHomePage),
        make(web.StandardFields),
}

// Get the singleton homePage object.
func GetHomePage() web.ViewModel { return web.ViewModel(home) }

func (h *homePage) MakeFields(app *web.App) (fields interface{}) {
        fields = h.StandardFields.MakeFields(app)
        names := []string{ "name1", "name2", "name3" }
        h.StandardFields["names"] = names
        return
}
