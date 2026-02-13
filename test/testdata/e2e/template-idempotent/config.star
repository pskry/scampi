target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="already rendered",
            content="hello",
            dest="/tmp/out.txt",
            perm="0644",
            owner="testuser",
            group="testgroup",
        ),
    ],
)
