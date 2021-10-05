version=0.1
for os in linux freebsd darwin; do
    for arch in amd64; do
        make clean
        make build GOOS=$os GOARCH=$arch

        pushd dist
        tar -czf ../juun-fzf-dist/juun-fzf-v$version-$os-$arch.tar.gz *
        pushd ../juun-fzf-dist/

        rm juun-fzf-latest-$os-$arch.tar.gz
        cp juun-fzf-v$version-$os-$arch.tar.gz juun-latest-$os-$arch.tar.gz
        shasum -a 256 juun-latest-$os-$arch.tar.gz
        popd
        popd
    done
done
