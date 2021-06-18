source "qingcloud-my-builder" "basic-example" {
  mock = "mock-config"
}

build {
  sources = [
    "source.qingcloud-my-builder.basic-example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo build generated data: ${build.GeneratedMockData}",
    ]
  }
}
