load("@npm_bazel_typescript//:index.bzl", "ts_library")

package(default_visibility = ["//visibility:public"])

ts_library(
    name = "terminal",
    srcs = glob(["*.tsx"]),
    deps = [
        "@npm//@types/react",
        "@npm//react",
        "@npm//xterm",
        "@npm//xterm-addon-fit",
    ],
)
