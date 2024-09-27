package swagger

import "embed"

//go:embed swagger.yaml
var Swaggerfile embed.FS
