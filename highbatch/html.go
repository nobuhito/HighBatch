package highbatch

import (
	"strings"
)

func getHtml(jsCommand string) string {
	return strings.Replace(html, "%%JSCOMMAND%%", jsCommand, -1)
}

const html = `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>HighBatch</title>
    <meta charset="utf-8">
    <meta name="description" content="">
    <meta name="author" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.1/jquery.min.js"></script>
    <script src="//maxcdn.bootstrapcdn.com/bootstrap/3.3.4/js/bootstrap.min.js"></script>
    <link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css">
    <link rel="stylesheet" href="/css/zenburn.css">
    <link rel="stylesheet" href="/css/highbatch.css">
    <link rel="shortcut icon" href="">
</head>
<body>
    <!-- Place your content here -->
    <nav class="navbar navbar-default navbar-fixed-top">
        <div class="container">
            <div class="navbar-header">
                <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar-link">
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                </button>
                <a class="navbar-brand" href="/">
                    HighBatch
                </a>
            </div>
            <div class="collapse navbar-collapse" id="navbar-link">
                <ul class="nav navbar-nav">
                    <li id="nav_tasks">
                        <a href="/tasks">
                            <span class="glyphicon glyphicon-time"></span>
                            Takss
                        </a>
                    </li>
                    <li id="nav_conf">
                        <a href="/conf">
                            <span class="glyphicon glyphicon-cog"></span>
                            Conf
                        </a>
                    </li>
                </ul>
            </div>
        </div>
    </nav>
    <div class="container">
        <div class="row" id="main">
            <div class="col-sm-8" id="article-wrap">
                <article id="article"></article>
            </div>
            <div class="col-sm-4" id="nav-wrap" >
                <nav id="nav"></nav>
            </div>
        </div>
        <div id="lightbox-outer"></div>
    </div>
    <!-- SCRIPTS -->
    <script src="/js/highlight.pack.js"></script>
    <script src="/js/bootstrap-treeview.min.js"></script>
    <script src="/js/highbatch.js"></script>
    <script>
       %%JSCOMMAND%%
    </script>
</body>
</html>
`
