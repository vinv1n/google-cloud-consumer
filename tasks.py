from invoke import task


@task(help={
    'tag': "Tag of the container"
})
def build(c, tag="consumer", cache=False):
    """
    Builds google cloud consumer container
    """
    if cache:
        c.run(f"docker build -f docker/Dockerfile -t {tag} $(pwd)")
    else:
        c.run(f"docker build -f docker/Dockerfile -t {tag} $(pwd) --no-cache")

@task(help={
    "tag": "tag or id of container",
    "deamon": "Run deamon as a deamon (default False)"
})
def run(c, tag="consumer", deamon=False):
    """
    Run google cloud consumer container
    """
    if deamon:
        c.run(f"docker run --rm -d -t {tag} -p 8080:8080")
    else:
        c.run(f"docker run --rm -t {tag} -p 8080:8080")
