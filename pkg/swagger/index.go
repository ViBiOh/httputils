package swagger

var (
	index = []byte(`
<!doctype html>
<html class="no-js" lang="">
	<head>
		<meta charset="utf-8">
		<title>Swagger UI</title>
		<meta name="description" content="Swagger UI">
		<meta name="viewport" content="width=device-width, initial-scale=1">

		<link rel="stylesheet" type="text/css" href="//unpkg.com/swagger-ui-dist@3/swagger-ui.css">
	</head>

	<body>
		<div id="swagger-ui"></div>

		<script src="//unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
		<script>
			SwaggerUIBundle({
			url: "./swagger.yaml",
			dom_id: '#swagger-ui',
			presets: [SwaggerUIBundle.presets.apis]
			})
		</script>
	</body>
</html>`)
)
