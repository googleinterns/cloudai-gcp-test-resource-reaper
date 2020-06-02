# load http_archive rule
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# load git_repository rule
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

# download go bazel tools
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "6a68e269802911fa419abb940c850734086869d7fe9bc8e12aaf60a09641c818",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.0/rules_go-v0.23.0.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.0/rules_go-v0.23.0.tar.gz",
    ],
)

# download gazelle
http_archive(
    name = "bazel_gazelle",
    sha256 = "bfd86b3cbe855d6c16c6fce60d76bd51f5c8dbc9cfcaef7a2bb5c1aafd0710e8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.0/bazel-gazelle-v0.21.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.0/bazel-gazelle-v0.21.0.tar.gz",
    ],
)

# load go rules
load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

# load gazelle rules
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
gazelle_dependencies()

# load proto rules
git_repository(
    name = "com_google_protobuf",
    commit = "09745575a923640154bcf307fba8aedff47f240a",
    remote = "https://github.com/protocolbuffers/protobuf",
    shallow_since = "1558721209 -0700",
)
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
protobuf_deps()
