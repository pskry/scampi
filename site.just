hugo := "go tool github.com/gohugoio/hugo"
mdlint := "go tool github.com/shinagawa-web/gomarklint"

[default]
[doc("Build the Hugo site")]
build:
  cp scampi.png site/static/scampi.png
  cd site && {{hugo}} --minify

[doc("Local dev server")]
dev *args:
  cp scampi.png site/static/scampi.png
  cd site && {{hugo}} server -D -e production {{args}}

[doc("Package site for deployment")]
dist outdir="build/site-dist": build
  mkdir -p {{outdir}}
  tar czf {{outdir}}/site.tar.gz -C build/site .

[doc("Lint markdown content")]
lint:
  cd site && {{mdlint}} content/

[doc("Remove all generated files")]
clean:
  rm -rf build/site build/site-dist site/resources/_gen
