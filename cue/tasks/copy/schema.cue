package copy

#Task: {
  kind: string

  if kind == "copy" {
    copy: {
      src: string
      dest: string
      perm: string
      owner: string
      group: string
    }
  }
}
