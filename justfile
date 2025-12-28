bld_dir := "./build"
bin_dir := f"{{bld_dir}}/bin"
bin_path := f"{{bin_dir}}/doit"

[default]
[doc("Show this help message")]
@help:
  just --list

[doc("Build the doit CLI binary")]
@build:
  mkdir -p {{bin_dir}}
  go build -o {{bin_path}} ./cmd

[doc("Build and run doit locally")]
@doit *args:
  go run ./cmd {{args}}

[doc("Clean project")]
@clean:
  rm -rf {{bld_dir}}
  go clean -testcache
