package evaluator

import obj "github.com/theawakener0/Zod/object"

var moduleCache = map[string]*obj.Enviroment{}
var loadingModules = map[string]bool{}
