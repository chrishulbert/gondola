package main

const (
	indexHtml = "index.html"
	htmlStart = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Gondola</title>
	<style>
	* {
		margin:0;
		padding:0;
		border:0;
		outline:0;
		font-family: 'Iowan Old Style', 'Palatino Linotype', 'URW Palladio L', P052, serif;
		border-collapse: collapse;
	}
	h1, h3 {
		padding: 0.3em;
		text-align: center;
	}
	table {
		width: 100%;
	}
	td {
		vertical-align: top;
		width: 50%;
		padding: 0.5em;
	}
	p {
		text-align: center;
		width: 100%;
	}
	img {
		width: 9em;
		height: 9em;
		object-fit: contain;
		background-color: gray;
		background-size: 1000%;
		background-position: center;
	}
	a {
		text-decoration: none;
	}
	</style>
</head>

<body>

	<h1>Gondola</h1>

	<table>`
	htmlTd = `<td>
				<a href="LINK">
					<p><img src="IMAGE" style='background-image: url("IMAGE")' /></p>
					<p>NAME</p>
				</a>
			</td>`
	htmlEnd = `
	</table>

	<h3>
		Gondola made with &hearts; by <a href="http://www.splinter.com.au">Chris Hulbert</a>
		<br />
		<a href="http://www.splinter.com.au/gondola">splinter.com.au/gondola</a>
		<br />
		<small><a href="javascript:alert('CAPACITY')">Usage</a></small>
		<br />
		<br />
		<small><em>This product uses the <a href="https://www.themoviedb.org">TMDb</a> API but is not endorsed or certified by TMDb.</em></small>
	</h3>

</body>
</html>`
)
