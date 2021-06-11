packer {
  required_plugins {
    qingcloud = {
      version = ">=v0.1.0"
      source  = "github.com/yunify/qingcloud"
    }
  }
}

source "qingcloud-my-builder" "foo-example" {
  mock = local.foo
}

source "qingcloud-my-builder" "bar-example" {
  mock = local.bar
}

build {
  sources = [
    "source.qingcloud-my-builder.foo-example",
  ]

  source "source.qingcloud-my-builder.bar-example" {
    name = "bar"
  }

  provisioner "qingcloud-my-provisioner" {
    only = ["qingcloud-my-builder.foo-example"]
    mock = "foo: ${local.foo}"
  }

  provisioner "qingcloud-my-provisioner" {
    only = ["qingcloud-my-builder.bar"]
    mock = "bar: ${local.bar}"
  }

  post-processor "qingcloud-my-post-processor" {
    mock = "post-processor mock-config"
  }
}
