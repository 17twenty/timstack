package flash

import "embed"

// This uses tailwind but you can replace this as required

//go:embed flashTemplate.tmpl.html
var f embed.FS
var flashTemplate, _ = f.ReadFile("flashTemplate.tmpl.html")
