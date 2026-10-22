package server

import (
	_ "embed"

	"github.com/gogf/gf/v2/net/ghttp"
)

// GoFrame 内置的 swagger 模板外链 https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js，
// 该 CDN 在国内常返回 200 且 0 字节的响应，脚本不执行、<redoc> 自定义元素无人注册，
// 页面表现为整页空白且控制台无报错；而 latest 构建已改为 ESM，普通 <script> 加载也不会自初始化。
// 故把 redoc v2.4.0 的 UMD 构建（末尾自执行并读取 spec-url）内嵌进库，由本服务直接吐出。

//go:embed assets/redoc.standalone.js
var redocStandaloneJS string

// redocAssetPath 内嵌 redoc 脚本的访问路径，与 swaggerUITemplate 中的 src 对应。
const redocAssetPath = "/doc-assets/redoc.standalone.js"

// swaggerUITemplate 中的 {SwaggerUIDocUrl} 由 GoFrame 输出时替换为 server.openapiPath。
const swaggerUITemplate = `
<!DOCTYPE html>
<html>
	<head>
		<title>API Reference</title>
		<meta charset="utf-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			body { margin: 0; padding: 0; }
		</style>
	</head>
	<body>
		<redoc spec-url="{SwaggerUIDocUrl}" show-object-schema-examples="true"></redoc>
		<script src="` + redocAssetPath + `"></script>
	</body>
</html>
`

// serveRedoc 注册内嵌 redoc 脚本的路由，供 swagger 页面加载。
func serveRedoc(s *ghttp.Server) {
	s.BindHandler(redocAssetPath, func(r *ghttp.Request) {
		r.Response.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		r.Response.Header().Set("Cache-Control", "public, max-age=86400")
		r.Response.Write(redocStandaloneJS)
	})
	s.SetSwaggerUITemplate(swaggerUITemplate)
}
